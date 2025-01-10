package main

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/mf751/blogman/ui"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.errLogger.Output(2, err.Error())
	app.render(w, "serverError", http.StatusInternalServerError, templateData{})
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, "notFound", http.StatusNotFound, templateData{})
}

func (app *application) render(
	w http.ResponseWriter,
	page string,
	statusCode int,
	data templateData,
) {
	if page == "serverError" {
		patterns := []string{"html/pages/serverError.html"}
		templateSet, _ := template.ParseFS(ui.Files, patterns...)
		templateSet.ExecuteTemplate(w, "serverError", nil)
		w.WriteHeader(statusCode)
		return
	} else if page == "notFound" {
		patterns := []string{"html/pages/notFound.html"}
		templateSet, _ := template.ParseFS(ui.Files, patterns...)
		templateSet.ExecuteTemplate(w, "notFound", nil)
		w.WriteHeader(statusCode)
		return
	}
	templateSet, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := templateSet.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
}
