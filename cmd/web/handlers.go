package main

import (
	"net/http"
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
		users = append(users, &user)
	}

	data := app.newTemplateData(r)
	data.Blogs = blogs
	data.Users = users

	app.render(w, "home", http.StatusInternalServerError, data)
}
