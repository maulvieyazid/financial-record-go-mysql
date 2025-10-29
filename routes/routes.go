package routes

import (
	"database/sql"
	"financial-record/config"
	"financial-record/controllers"
	"net/http"
)

func Routes(db *sql.DB) {

	authController := controllers.NewAuthController(db)
	http.HandleFunc("/register", config.GuestOnly(authController.Register))
	http.HandleFunc("/login", config.GuestOnly(authController.Login))
	http.HandleFunc("/logout", config.AuthOnly(authController.Logout))

	financialController := controllers.NewFinancialController(db)
	http.HandleFunc("/", config.AuthOnly(financialController.Home))
	http.HandleFunc("/home", config.AuthOnly(financialController.Home))
	http.HandleFunc("/financial/add_financial_record", config.AuthOnly(financialController.AddFinacialRecord))
	http.HandleFunc("/financial/delete_financial_record", config.AuthOnly(financialController.DeleteFinancialRecord))
	http.HandleFunc("/financial/download_financial_record", config.AuthOnly(financialController.DownloadFinancialRecord))
	http.HandleFunc("/financial/edit_financial_record", config.AuthOnly(financialController.EditFinancialRecord))

	userController := controllers.NewUserController(db)
	http.HandleFunc("/profile", config.AuthOnly(userController.Profile))
}
