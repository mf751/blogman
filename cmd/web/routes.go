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
	mux.Handle(http.MethodGet+" /user/{username}", firstLayer.ThenFunc(app.userBlogs))
	mux.Handle(http.MethodGet+" /search", firstLayer.ThenFunc(app.search))

	secondLayer := firstLayer.Append(app.requireAuthentication)
	mux.Handle(http.MethodGet+" /blogs", secondLayer.ThenFunc(app.myBlogs))
	mux.Handle(http.MethodGet+" /blog/create", secondLayer.ThenFunc(app.blogCreate))
	mux.Handle(http.MethodPost+" /blog/create", secondLayer.ThenFunc(app.blogCreatePost))
	mux.Handle(http.MethodGet+" /blog/update/{id}", secondLayer.ThenFunc(app.blogUpdate))
	mux.Handle(http.MethodPost+" /blog/update", secondLayer.ThenFunc(app.blogUpdatePost))
	mux.Handle(http.MethodPost+" /blog/delete", secondLayer.ThenFunc(app.blogDeletePost))
	mux.Handle(http.MethodGet+" /account", secondLayer.ThenFunc(app.userAccount))
	mux.Handle(http.MethodGet+" /password/change", secondLayer.ThenFunc(app.userChangePassword))
	mux.Handle(
		http.MethodPost+" /password/change",
		secondLayer.ThenFunc(app.userChangePasswordPost),
	)
	mux.Handle(http.MethodPost+" /user/logout", secondLayer.ThenFunc(app.userLogoutPost))

	// not found
	mux.Handle("GET /", firstLayer.ThenFunc(app.notFound))

	bottomLayer := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return bottomLayer.Then(mux)
}
