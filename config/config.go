package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// InitConfiguration loads configuration for the application.
// It will first try to load `.env` (if present) and then try to read
// `app.conf.json`. Environment variables override values from the
// configuration file. Use environment variables in production or
// provide a `.env` file for local development.
func InitConfiguration() {
	// Load .env if present (non-fatal)
	_ = godotenv.Load()

	viper.SetConfigName("app.conf")
	viper.SetConfigType("json")
	// Look in multiple locations so tests running from package folders
	// can still find the project-level `app.conf.json`.
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("app.conf.json not found, continuing with environment variables/.env")
	}

	// Allow environment variables to override config values.
	// Convert env names like DATABASE_USER to viper key DATABASE.USER
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// sensible defaults for local development/tests
	viper.SetDefault("DATABASE.DRIVER", "mysql")
}
