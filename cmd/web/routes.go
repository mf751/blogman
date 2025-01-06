package main

import (
	"net/http"

	"github.com/justinas/alice"

	"github.com/mf751/blogman/ui"
)

func (app *application) mainMux() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(ui.Files))
	mux.Handle("GET /static/", fileServer)

	firstLayer := alice.New(secureHeaders)

	mux.Handle(http.MethodGet+" /{$}", firstLayer.ThenFunc(app.home))

	mux.Handle("GET /", firstLayer.ThenFunc(app.notFound))

	return firstLayer.Then(mux)
}
