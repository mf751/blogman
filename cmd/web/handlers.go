package main

import "net/http"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	_, err := app.blogs.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// data := app.newTemplateData(r)
	// data.blogs := blogs
	data := templateData{}
	app.render(w, "home", http.StatusInternalServerError, data)
}
