package config

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfiguration() {
	viper.SetConfigName("app.conf")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Konfigurasi error, ", err)
	}
}
