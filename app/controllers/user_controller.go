package controllers

import (
	"database/sql"
	"financial-record/config"
	"financial-record/entities"
	"financial-record/helpers"
	"financial-record/models"
	"financial-record/views"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	db *sql.DB
}

func NewUserController(db *sql.DB) *UserController {
	return &UserController{
		db: db,
	}
}

func (controller UserController) Profile(writer http.ResponseWriter, request *http.Request) {

	templateLayout := "views/user/profile.html"

	// untuk mengirim data ke html
	var data = make(map[string]interface{})

	// panggil session
	sessions, _ := config.Store.Get(request, config.SESSION_ID)

	// tampilkan alert dari session
	if flashes := sessions.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		sessions.Save(request, writer)
	}

	// ambil user id dari session
	sessionUserId := sessions.Values["ID"].(string)

	// tampilkan data user berdasarkan id
	user, err := models.NewUserModel(controller.db).FindUserById(sessionUserId)
	if err != nil {
		data["error"] = "User tidak ditemukan, " + err.Error()
	} else {
		data["user"] = user
	}

	if request.Method == http.MethodPost {

		request.ParseMultipartForm(5 * 1024 * 1024) // file maksimal 5MB

		password := request.Form.Get("password")
		user := entities.User{
			Id:    sessionUserId,
			Name:  request.Form.Get("name"),
			Email: request.Form.Get("email"),
		}

		// tampilkan error sesuai ketentuan di Struct
		if err := helpers.NewValidator(controller.db).Struct(user); err != nil {
			data["validation"] = err
			data["user"] = user
			views.RenderTemplate(writer, templateLayout, data)
			return
		}

		// kalau user ganti password, hash password baru
		if password != "" {
			hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			user.Password = string(hashPassword)
		}

		if file, handler, err := request.FormFile("photo"); err == nil {
			defer file.Close() // jeda proses sampai dipilih filenya

			// validasi ukuran
			if handler.Size > 5*1024*1024 {
				data["error"] = "Ukuran file terlalu besar, maksimal 5MB"
				data["user"] = user
				views.RenderTemplate(writer, templateLayout, data)
				return
			}

			//validasi ekstensi file
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				data["error"] = "File tidak didukung"
				data["user"] = user
				views.RenderTemplate(writer, templateLayout, data)
				return
			}

			// ambil foto sebelumnya jika ada
			oldPhoto, err := models.NewUserModel(controller.db).GetUserPhotoById(sessionUserId)
			if err != nil {
				data["error"] = "Gagal mengambil foto, " + err.Error()
				data["user"] = user
				views.RenderTemplate(writer, templateLayout, data)
				return
			}

			// atur foto baru ke folder public/user_photo
			filename := fmt.Sprintf("profile_%s%s", time.Now().Format("2006-01-02_15-04-05"), ext)
			path := filepath.Join("public/user_photo", filename)

			// hapus foto sebelumnya jika ada
			if oldPhoto != nil && *oldPhoto != "" {
				oldPath := filepath.Join("public/user_photo", *oldPhoto)
				if err := os.Remove(oldPath); err != nil {
					data["error"] = "Gagal menghapus foto sebelumnya, " + err.Error()
					data["user"] = user
					views.RenderTemplate(writer, templateLayout, data)
					return
				}
			}

			// buat file/foto baru
			out, err := os.Create(path)
			if err != nil {
				data["error"] = "Gagal membuat foto baru, " + err.Error()
				data["user"] = user
				views.RenderTemplate(writer, templateLayout, data)
				return
			}
			defer out.Close()

			// simpan file baru
			_, errCopy := io.Copy(out, file)
			if errCopy != nil {
				data["error"] = "Gagal menyimpan foto baru, " + errCopy.Error()
				data["user"] = user
				views.RenderTemplate(writer, templateLayout, data)
				return
			}

			// set ke struct
			user.Photo = &filename
		} else {
			// kalau user tidak ganti foto, pakai foto lama
			oldPhoto, err := models.NewUserModel(controller.db).GetUserPhotoById(sessionUserId)
			if err == nil {
				user.Photo = oldPhoto
			}
		}

		// update profile
		err := models.NewUserModel(controller.db).UpdateProfile(user)
		if err != nil {
			data["error"] = "Gagal mengubah data profile, " + err.Error()
		} else {
			sessions.AddFlash("Berhasil mengubah data profile", "success")
			sessions.Save(request, writer)
			http.Redirect(writer, request, "/profile", http.StatusSeeOther)
			return
		}
	}

	views.RenderTemplate(writer, templateLayout, data)
}
