package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLogger.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}
		ID, err := uuid.Parse(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		_, err = app.users.Get(ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), isAuthenticatedKey, id)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
