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
	mux.HandleFunc(http.MethodGet+" /ping", pong)

	firstLayer := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.Handle(http.MethodGet+" /{$}", firstLayer.ThenFunc(app.home))
	mux.Handle(http.MethodGet+" /about", firstLayer.ThenFunc(app.about))
	mux.Handle(http.MethodGet+" /blog/{id}", firstLayer.ThenFunc(app.blogView))

	// not found
	mux.Handle("GET /", firstLayer.ThenFunc(app.notFound))

	bottomLayer := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return bottomLayer.Then(mux)
}
