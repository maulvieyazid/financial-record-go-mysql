package main

import (
	"financial-record/config"
	"financial-record/routes"
	"log"
	"net/http"
)

func main() {

	// read file dari folder public
	http.Handle("/user_photo/", http.StripPrefix("/user_photo/", http.FileServer(http.Dir("public/user_photo"))))

	config.InitConfiguration()

	db := config.InitDatabase()
	routes.Routes(db)

	log.Println("Service berjalan di port :8000")
	http.ListenAndServe(":8000", nil)

}
