package views

import (
	"html/template"
	"net/http"
)

// RenderTemplate is a function variable so tests can replace it.
var RenderTemplate = func(writer http.ResponseWriter, path string, data interface{}) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(writer, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(writer, data)
}
