package views

import (
	"html/template"
	"net/http"
)

// fungsi untuk me-render template HTML
func RenderTemplate(writer http.ResponseWriter, path string, data interface{}) {
	template, err := template.ParseFiles(path)
	if err != nil {
		http.Error(writer, "Template error", http.StatusInternalServerError)
		return
	}
	template.Execute(writer, data)
}
