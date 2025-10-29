package config

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func InitDatabase() *sql.DB {

	dbUser := viper.GetString("DATABASE.USER")
	dbPass := viper.GetString("DATABASE.PASSWORD")
	dbName := viper.GetString("DATABASE.NAME")
	dbHost := viper.GetString("DATABASE.HOST")
	dbPort := viper.GetString("DATABASE.PORT")
	dbDriver := viper.GetString("DATABASE.DRIVER")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true&loc=Asia%2FJakarta"

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal("Gagal koneksi ke database", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Ping error", err)
	}

	log.Println("Koneksi ke database berhasil")
	return db
}
