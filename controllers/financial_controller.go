package controllers

import (
	"database/sql"
	"financial-record/config"
	"financial-record/entities"
	"financial-record/helpers"
	"financial-record/models"
	"financial-record/views"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FinancalController struct {
	db *sql.DB
}

func NewFinancialController(db *sql.DB) *FinancalController {
	return &FinancalController{
		db: db,
	}
}

// format rupiah
func formatIDR(n int64) string {
	str := fmt.Sprintf("%d", n)
	var result []string
	for len(str) > 3 {
		result = append([]string{str[len(str)-3:]}, result...)
		str = str[:len(str)-3]
	}
	if len(str) > 0 {
		result = append([]string{str}, result...)
	}
	return strings.Join(result, ".") + ",00"
}

func (controller *FinancalController) Home(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/financial/home.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	// tampilkan alert dari session
	if flashes := sessions.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		sessions.Save(request, writer)
	}

	// tampilkan dropdown bulan
	currentDate := time.Now()
	var months []string
	for i := 0; i < 6; i++ {
		previousMonth := currentDate.AddDate(0, -i, 0)
		months = append(months, previousMonth.Format("January 2006"))
	}
	data["months"] = months

	// trigger ketika dropdown bulan dipilih
	selectedMonth := request.URL.Query().Get("selected_month")
	if selectedMonth == "" {
		selectedMonth = currentDate.Format("January 2006")
	}
	data["selectedMonth"] = selectedMonth

	// trigger ketika checkbox Pemasukan/Pengeluaran dipilih
	pemasukanOnly := request.URL.Query().Get("pemasukanOnly") == "true"
	data["pemasukanOnly"] = pemasukanOnly
	pengeluaranOnly := request.URL.Query().Get("pengeluaranOnly") == "true"
	data["pengeluaranOnly"] = pengeluaranOnly

	// tampilkan total Pemasukan dan Pengeluaran
	sessionUserId := sessions.Values["ID"].(string)
	model := models.NewFinancalModel(controller.db)
	totalPemasukan, totalPengeluaran, err := model.GetFinancialTotalNominal(sessionUserId, selectedMonth, pemasukanOnly, pengeluaranOnly)
	if err != nil {
		data["error"] = "Gagal mendapatkan total data keuangan, " + err.Error()
	} else {
		data["total_pemasukan"] = totalPemasukan
		data["total_pengeluaran"] = totalPengeluaran
	}

	// tampilkan list keuangan
	financials, err := model.FindAllFinancial(sessionUserId, selectedMonth, pemasukanOnly, pengeluaranOnly)
	if err != nil {
		data["error"] = "Gagal menampilkan list data keuangan, " + err.Error()
	} else {
		data["financials"] = financials
	}

	funcMap := template.FuncMap{
		"formatIDR": formatIDR,
		"indexNo":   func(a, b int) int { return a + b },
	}

	template, _ := template.New(filepath.Base(templateLayout)).Funcs(funcMap).ParseFiles(templateLayout)
	template.Execute(writer, data)
}

func (controller *FinancalController) AddFinacialRecord(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/financial/create.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// panggil session
	session, _ := config.Store.Get(request, config.SESSION_ID)

	if request.Method == http.MethodPost {

		request.ParseForm()

		// ambil tanggal
		dateStr := request.Form.Get("date")
		date, _ := time.Parse("2006-01-02", dateStr)

		// ambil nominal
		nominalStr := request.Form.Get("nominal")
		nominal, _ := strconv.ParseInt(nominalStr, 10, 64)

		// ambil attachment
		var attachment *string
		if attachmentValue := request.Form.Get("attachment"); attachmentValue != "" {
			attachment = &attachmentValue
		}

		// ambil deskripsi
		var description *string
		if descriptionValue := request.Form.Get("description"); descriptionValue != "" {
			description = &descriptionValue
		}

		// ambil user id
		sessionUserId, _ := session.Values["ID"].(string)

		// masukkan ke struct
		financial := entities.AddFinancial{
			UserId:      sessionUserId,
			Date:        date,
			Type:        request.Form.Get("type"),
			Category:    request.Form.Get("category"),
			Nominal:     nominal,
			Description: description,
			Attachment:  attachment,
		}

		// tampilkan error sesuai ketentuan di Struct
		if err := helpers.NewValidator(controller.db).Struct(financial); err != nil {
			data["validation"] = err
			data["financial"] = financial
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// insert ke database
		if err := models.NewFinancalModel(controller.db).AddFinacialRecord(financial); err != nil {
			data["error"] = "Gagal menambahkan data keuangan, " + err.Error()
			views.RenderTemplate(writer, templateLayout, data)
			return
		} else {
			session.AddFlash("Berhasil menambahkan data keuangan", "success")
			session.Save(request, writer)
			http.Redirect(writer, request, "/home", http.StatusSeeOther)
			return
		}
	}

	views.RenderTemplate(writer, templateLayout, data)

}

func (controller *FinancalController) DeleteFinancialRecord(writer http.ResponseWriter, request *http.Request) {

	// panggil session
	session, _ := config.Store.Get(request, config.SESSION_ID)

	idStr := request.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 16)
	if idStr == "" || err != nil {
		session.AddFlash("Gagal mengambil data keuangan", "error")
		session.Save(request, writer)
		http.Redirect(writer, request, "/home", http.StatusSeeOther)
		return
	}

	if err := models.NewFinancalModel(controller.db).DeleteFinancialRecord(int16(id)); err != nil {
		session.AddFlash("Gagal menghapus data keuangan, "+err.Error(), "error")
		session.Save(request, writer)
	} else {
		session.AddFlash("Berhasil menghapus data keuangan", "success")
		session.Save(request, writer)
	}

	http.Redirect(writer, request, "/home", http.StatusSeeOther)
}

func (controller *FinancalController) DownloadFinancialRecord(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/financial/download.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	// tampilkan bulan yang dipilih
	selectedMonth := request.URL.Query().Get("selected_month")
	data["selectedMonth"] = selectedMonth

	// tampilkan data Pemasukan/Pengeluaran yang dipilih
	pemasukanOnly := request.URL.Query().Get("pemasukanOnly") == "true"
	data["pemasukanOnly"] = pemasukanOnly
	pengeluaranOnly := request.URL.Query().Get("pengeluaranOnly") == "true"
	data["pengeluaranOnly"] = pengeluaranOnly

	// tampilkan total Pemasukan dan Pengeluaran
	sessionUserId := sessions.Values["ID"].(string)
	model := models.NewFinancalModel(controller.db)
	totalPemasukan, totalPengeluaran, err := model.GetFinancialTotalNominal(sessionUserId, selectedMonth, pemasukanOnly, pengeluaranOnly)
	if err != nil {
		data["error"] = "Gagal mendapatkan total data keuangan, " + err.Error()
	} else {
		data["total_pemasukan"] = totalPemasukan
		data["total_pengeluaran"] = totalPengeluaran
	}

	// tampilkan list keuangan
	financials, err := model.FindAllFinancial(sessionUserId, selectedMonth, pemasukanOnly, pengeluaranOnly)
	if err != nil {
		data["error"] = "Gagal menampilkan list data keuangan, " + err.Error()
	} else {
		data["financials"] = financials
	}

	funcMap := template.FuncMap{
		"formatIDR": formatIDR,
		"indexNo":   func(a, b int) int { return a + b },
	}

	template, _ := template.New(filepath.Base(templateLayout)).Funcs(funcMap).ParseFiles(templateLayout)
	template.Execute(writer, data)

}

func (controller *FinancalController) EditFinancialRecord(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/financial/edit.html"

	// panggil session
	session, _ := config.Store.Get(request, config.SESSION_ID)

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// ambil id dari url
	idStr := request.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 16)
	if idStr == "" || err != nil {
		data["error"] = "Gagal mengambil data keuangan, " + err.Error()
		views.RenderTemplate(writer, templateLayout, data)
		return
	}

	// tampilkan data berdasarkan id
	findFinancial, err := models.NewFinancalModel(controller.db).FindFinancialById(int16(id))
	if err != nil {
		data["error"] = "Data keuangan tidak ditemukan, " + err.Error()
	} else {
		data["financial"] = findFinancial
	}

	if request.Method == http.MethodPost {

		request.ParseForm()

		// ambil tanggal
		dateStr := request.Form.Get("date")
		date, _ := time.Parse("2006-01-02", dateStr)

		// ambil nominal
		nominalStr := request.Form.Get("nominal")
		nominal, _ := strconv.ParseInt(nominalStr, 10, 64)

		// ambil attachment
		var attachment *string
		if attachmentValue := request.Form.Get("attachment"); attachmentValue != "" {
			attachment = &attachmentValue
		}

		// ambil deskripsi
		var description *string
		if descriptionValue := request.Form.Get("description"); descriptionValue != "" {
			description = &descriptionValue
		}

		// ambil user id
		sessionUserId, _ := session.Values["ID"].(string)

		// masukkan ke struct
		financial := entities.AddFinancial{
			Id:          int16(id),
			UserId:      sessionUserId,
			Date:        date,
			Type:        request.Form.Get("type"),
			Category:    request.Form.Get("category"),
			Nominal:     nominal,
			Description: description,
			Attachment:  attachment,
		}

		// tampilkan error sesuai ketentuan di Struct
		if err := helpers.NewValidator(controller.db).Struct(financial); err != nil {
			data["validation"] = err
			data["financial"] = financial
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// update data di database
		if err := models.NewFinancalModel(controller.db).EditFinancialRecord(financial); err != nil {
			data["error"] = "Gagal mengubah data keuangan, " + err.Error()
		} else {
			session.AddFlash("Berhasil mengubah data keuangan", "success")
			session.Save(request, writer)
			http.Redirect(writer, request, "/home", http.StatusSeeOther)
			return
		}

	}

	views.RenderTemplate(writer, templateLayout, data)
}
