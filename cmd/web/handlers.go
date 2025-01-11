package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/mf751/blogman/interanl/models"
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

	data := app.newTemplateData(r)
	data.Blog = blog

	app.render(w, "blog", http.StatusOK, data)
}
