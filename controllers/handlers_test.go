package controllers

import (
    "net/http"
    "net/http/httptest"
    "regexp"
    "testing"

    "financial-record/config"

    "github.com/DATA-DOG/go-sqlmock"
)

// helper: create request with optional logged-in session
func reqWithSession(t *testing.T, method, target string, loggedIn bool, id string) *http.Request {
    t.Helper()
    // initial request to get a session cookie
    r1 := httptest.NewRequest(method, target, nil)
    rr1 := httptest.NewRecorder()
    s, _ := config.Store.Get(r1, config.SESSION_ID)
    if loggedIn {
        s.Values["LOGGED_IN"] = true
        s.Values["ID"] = id
    }
    if err := s.Save(r1, rr1); err != nil {
        t.Fatalf("saving session failed: %v", err)
    }
    cookie := rr1.Header().Get("Set-Cookie")

    // new request that contains cookie
    r2 := httptest.NewRequest(method, target, nil)
    if cookie != "" {
        r2.Header.Set("Cookie", cookie)
    }
    return r2
}

func TestAuth_RegisterAndLoginGuestPages(t *testing.T) {
    // /register (guest) should return 200 for guest
    req := reqWithSession(t, http.MethodGet, "/register", false, "")
    rr := httptest.NewRecorder()

    ac := NewAuthController(nil)
    // wrap with GuestOnly middleware
    handler := config.GuestOnly(ac.Register)
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected status 200 for /register GET, got %d", rr.Code)
    }

    // /login (guest) should return 200 for guest
    req2 := reqWithSession(t, http.MethodGet, "/login", false, "")
    rr2 := httptest.NewRecorder()
    handler2 := config.GuestOnly(ac.Login)
    handler2.ServeHTTP(rr2, req2)
    if rr2.Code != http.StatusOK {
        t.Fatalf("expected status 200 for /login GET, got %d", rr2.Code)
    }
}

func TestAuth_Logout_RedirectsForLoggedIn(t *testing.T) {
    req := reqWithSession(t, http.MethodGet, "/logout", true, "userid-test")
    rr := httptest.NewRecorder()
    ac := NewAuthController(nil)
    handler := config.AuthOnly(ac.Logout)
    handler.ServeHTTP(rr, req)
    // Logout performs redirect to /login
    if rr.Code != http.StatusSeeOther {
        t.Fatalf("expected redirect status for /logout, got %d", rr.Code)
    }
}

func TestFinancial_HomeAndDownload_WithMockDB(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    defer db.Close()

    // Expectation for GetFinancialTotalNominal -> QueryRow returning 0,0
    mock.ExpectQuery("COALESCE\\(SUM").WillReturnRows(sqlmock.NewRows([]string{"total_pemasukan", "total_pengeluaran"}).AddRow(0, 0))

    // Expectation for FindAllFinancial -> return empty rows
    mock.ExpectQuery("SELECT id, date, type, category, nominal, description, attachment").WillReturnRows(sqlmock.NewRows([]string{"id", "date", "type", "category", "nominal", "description", "attachment"}))

    fc := NewFinancialController(db)

    req := reqWithSession(t, http.MethodGet, "/", true, "userid-test")
    rr := httptest.NewRecorder()
    handler := config.AuthOnly(fc.Home)
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected status 200 for Home GET, got %d", rr.Code)
    }

    // Download page (GET) - similar expectations
    mock.ExpectQuery("COALESCE\\(SUM").WillReturnRows(sqlmock.NewRows([]string{"total_pemasukan", "total_pengeluaran"}).AddRow(0, 0))
    mock.ExpectQuery("SELECT id, date, type, category, nominal, description, attachment").WillReturnRows(sqlmock.NewRows([]string{"id", "date", "type", "category", "nominal", "description", "attachment"}))

    req2 := reqWithSession(t, http.MethodGet, "/financial/download_financial_record", true, "userid-test")
    rr2 := httptest.NewRecorder()
    handler2 := config.AuthOnly(fc.DownloadFinancialRecord)
    handler2.ServeHTTP(rr2, req2)
    if rr2.Code != http.StatusOK {
        t.Fatalf("expected status 200 for DownloadFinancialRecord GET, got %d", rr2.Code)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet sqlmock expectations: %v", err)
    }
}

func TestFinancial_AddEditDelete_WithMockDB(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    defer db.Close()

    fc := NewFinancialController(db)

    // Add (GET) should just render template, no DB calls needed
    reqAdd := reqWithSession(t, http.MethodGet, "/financial/add_financial_record", true, "userid-test")
    rrAdd := httptest.NewRecorder()
    handlerAdd := config.AuthOnly(fc.AddFinacialRecord)
    handlerAdd.ServeHTTP(rrAdd, reqAdd)
    if rrAdd.Code != http.StatusOK {
        t.Fatalf("expected status 200 for AddFinacialRecord GET, got %d", rrAdd.Code)
    }

    // Edit (GET) expects FindFinancialById -> return one row
    mock.ExpectQuery("SELECT id, date, type, category, nominal, description, attachment FROM record WHERE id = ?").WillReturnRows(
        sqlmock.NewRows([]string{"id", "date", "type", "category", "nominal", "description", "attachment"}).AddRow(1, "2020-01-01", "pemasukan", "kat", 10000, "desc", "attach.jpg"))

    // create URL with id param
    u := "/financial/edit_financial_record?id=1"
    reqEdit := reqWithSession(t, http.MethodGet, u, true, "userid-test")
    rrEdit := httptest.NewRecorder()
    handlerEdit := config.AuthOnly(fc.EditFinancialRecord)
    handlerEdit.ServeHTTP(rrEdit, reqEdit)
    if rrEdit.Code != http.StatusOK {
        t.Fatalf("expected status 200 for EditFinancialRecord GET, got %d", rrEdit.Code)
    }

    // Delete (with id) expects Exec
    mock.ExpectExec("DELETE FROM record").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
    reqDel := reqWithSession(t, http.MethodGet, "/financial/delete_financial_record?id=1", true, "userid-test")
    rrDel := httptest.NewRecorder()
    handlerDel := config.AuthOnly(fc.DeleteFinancialRecord)
    handlerDel.ServeHTTP(rrDel, reqDel)
    // handler redirects to /home after deletion
    if rrDel.Code != http.StatusSeeOther {
        t.Fatalf("expected redirect for DeleteFinancialRecord, got %d", rrDel.Code)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet sqlmock expectations: %v", err)
    }
}

func TestUser_Profile_Get_WithMockDB(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    defer db.Close()

    // Expect FindUserById
    mock.ExpectQuery(regexp.QuoteMeta("SELECT email, name, photo FROM users WHERE id = ?")).WithArgs("userid-test").WillReturnRows(
        sqlmock.NewRows([]string{"email", "name", "photo"}).AddRow("a@b.com", "Test User", "photo.jpg"))

    uc := NewUserController(db)
    req := reqWithSession(t, http.MethodGet, "/profile", true, "userid-test")
    rr := httptest.NewRecorder()
    handler := config.AuthOnly(uc.Profile)
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected status 200 for Profile GET, got %d", rr.Code)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet sqlmock expectations: %v", err)
    }
}
