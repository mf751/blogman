package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/mf751/blogman/interanl/models"
	"github.com/mf751/blogman/interanl/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	blogs, err := app.blogs.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	users := []*models.User{}
	for _, blog := range blogs {
		blog.Content = blog.Content[:500] + "..."
		user, err := app.users.Get(blog.UserID)
		user.Created = time.Time{}
		user.Email = ""
		if err != nil {
			app.serverError(w, err)
			return
		}
		users = append(users, user)
	}

	data := app.newTemplateData(r)
	data.Blogs = blogs
	data.Users = users
	data.Active = "blogs"

	app.render(w, "home", http.StatusInternalServerError, data)
}

func pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
}

func (app *application) blogView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ID, err := strconv.Atoi(id)
	if err != nil || ID < 1 {
		app.notFound(w, r)
		return
	}
	blog, err := app.blogs.Get(ID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	user, err := app.users.Get(blog.UserID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Blog = blog
	data.User = user
	data.Active = "None"

	app.render(w, "blog", http.StatusOK, data)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	data.Active = "login"
	app.render(w, "login", http.StatusOK, data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		data.Active = "login"
		app.render(w, "login", http.StatusUnprocessableEntity, data)
		return
	}
	ID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrWrongCredintials) {
			form.AddNonFieldError("Email or Password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			data.Active = "login"
			app.render(w, "login", http.StatusUnprocessableEntity, data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", (*ID).String())
	originalUrl := app.sessionManager.GetString(r.Context(), "original-path")
	if originalUrl != "" {
		app.sessionManager.Remove(r.Context(), "original-path")
		http.Redirect(w, r, originalUrl, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out succussfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
