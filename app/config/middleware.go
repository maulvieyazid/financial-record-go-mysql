package config

import (
	"net/http"
)

func GuestOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, SESSION_ID)
		if session.Values["LOGGED_IN"] == true {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func AuthOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, SESSION_ID)
		if session.Values["LOGGED_IN"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
