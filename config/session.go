package config

import (
	"github.com/gorilla/sessions"
)

const SESSION_ID = "finacial_record_okt"
const FLASH_ID = "flash_logout"

var Store *sessions.CookieStore

func init() {
	Store = sessions.NewCookieStore([]byte(SESSION_ID))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // session akan disimpan selama 24 jam
		HttpOnly: true,
		Secure:   true,
	}
}
