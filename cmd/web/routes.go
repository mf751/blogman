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
	mux.Handle(http.MethodGet+" /user/login", firstLayer.ThenFunc(app.userLogin))
	mux.Handle(http.MethodPost+" /user/login", firstLayer.ThenFunc(app.userLoginPost))
	mux.Handle(http.MethodGet+" /user/signup", firstLayer.ThenFunc(app.userSignup))
	mux.Handle(http.MethodPost+" /user/signup", firstLayer.ThenFunc(app.userSignupPost))

	secondLayer := firstLayer.Append(app.requireAuthentication)
	mux.Handle(http.MethodGet+" /blog", secondLayer.ThenFunc(app.blogCreate))
	mux.Handle(http.MethodPost+" /blog", secondLayer.ThenFunc(app.blogCreatePost))
	mux.Handle(http.MethodPost+" /user/logout", secondLayer.ThenFunc(app.userLogoutPost))

	// not found
	mux.Handle("GET /", firstLayer.ThenFunc(app.notFound))

	bottomLayer := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return bottomLayer.Then(mux)
}
