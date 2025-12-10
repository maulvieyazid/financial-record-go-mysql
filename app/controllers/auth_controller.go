package controllers

import (
	"database/sql"
	"financial-record/config"
	"financial-record/entities"
	"financial-record/helpers"
	"financial-record/models"
	"financial-record/views"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	db *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		db: db,
	}
}

func (controller *AuthController) Register(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/auth/register.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// untuk mencegah <no value> di awal
	data["register"] = entities.Register{}

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	if request.Method == http.MethodPost {

		// kalau ambil data dari name -> request.ParseForm()
		request.ParseForm()
		register := entities.Register{
			Id:              uuid.New().String(),
			Name:            request.Form.Get("name"),
			Email:           request.Form.Get("email"),
			Password:        request.Form.Get("password"),
			ConfirmPassword: request.Form.Get("confirm_password"),
		}

		// tampilkan error dari validator
		if err := helpers.NewValidator(controller.db).Struct(register); err != nil {
			data["validation"] = err
			data["register"] = register
			fmt.Println(err)
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// ubah password yang diinput menjadi hash
		hashPassword, _ := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
		register.Password = string(hashPassword)

		// insert ke database
		if err := models.NewAuthModel(controller.db).Register(register); err != nil {
			data["error"] = "Registrasi gagal, error: " + err.Error()
		} else {
			data["success"] = "Registrasi berhasil, silahkan login"
		}

		// kirim pesan success dengan session ke halaman login
		sessions.AddFlash("Registrasi berhasil, silahkan login untuk melanjutkan", "success")
		sessions.Save(request, writer)

		// redirect ke login
		http.Redirect(writer, request, "/login", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(writer, templateLayout, data)
}

func (controller *AuthController) Login(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/auth/login.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	// tampilkan alert dari session
	flashSession, _ := config.Store.Get(request, config.FLASH_ID)
	if flashes := flashSession.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		flashSession.Save(request, writer)
	}

	// untuk mencegah <no value> di awal
	data["login"] = entities.Register{}

	if request.Method == http.MethodPost {

		// ambil inputan
		request.ParseForm()
		login := entities.Auth{
			Email:    request.Form.Get("email"),
			Password: request.Form.Get("password"),
		}

		// tampilkan error dari validator
		if err := helpers.NewValidator(controller.db).Struct(login); err != nil {
			data["validation"] = err
			data["login"] = login
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// cari berdasarkan email
		user, err := models.NewAuthModel(controller.db).Login(login.Email)
		if err != nil {
			data["error"] = "Akun tidak ditemukan"
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// cocokkan password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			data["error"] = "Password anda salah"
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// simpan data ke session
		sessions.Values["LOGGED_IN"] = true
		sessions.Values["ID"] = user.Id

		// tampilkan alert success di home
		sessions.AddFlash("Selamat datang "+user.Name, "success")
		sessions.Save(request, writer)

		// redirect ke home
		http.Redirect(writer, request, "/home", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(writer, templateLayout, data)
}

func (controller *AuthController) Logout(writer http.ResponseWriter, request *http.Request) {

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	// hapus semua data di session
	sessions.Options.MaxAge = -1
	sessions.Save(request, writer)

	// simpan pesan ke session terpisah
	flashSession, _ := config.Store.Get(request, config.FLASH_ID)
	flashSession.AddFlash("Berhasil logout", "success")
	flashSession.Save(request, writer)

	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}
