package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) mainMux() http.Handler {
	mux := httprouter.New()

	firstLayer := alice.New(secureHeaders)

	mux.Handler(http.MethodGet, "/", firstLayer.ThenFunc(app.home))

	return firstLayer.Then(mux)
}
