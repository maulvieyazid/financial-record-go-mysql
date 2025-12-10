package config

import (
	"os"
	"strings"

	"github.com/gorilla/sessions"
)

const SESSION_ID = "finacial_record_okt"
const FLASH_ID = "flash_logout"

var Store *sessions.CookieStore
func InitStore() {
	// Secure cookie default: false for local/dev. Enable by env var `APP_SECURE_COOKIE=true` in production (HTTPS).
	secure := false
	if strings.ToLower(os.Getenv("APP_SECURE_COOKIE")) == "true" || strings.ToLower(os.Getenv("APP_ENV")) == "production" {
		secure = true
	}

	Store = sessions.NewCookieStore([]byte(SESSION_ID))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // session akan disimpan selama 24 jam
		HttpOnly: true,
		Secure:   secure,
	}
}

func init() {
	InitStore()
}
