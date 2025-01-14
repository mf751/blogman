package main

import (
	"fmt"
	"html/template"
	"time"

	"github.com/mf751/blogman/internal/models"
	"github.com/mf751/blogman/ui"
)

type templateData struct {
	Blog            *models.Blog
	User            *models.User
	Blogs           []*models.Blog
	Users           []*models.User
	Form            any
	Active          string
	CSRFToken       string
	IsAuthenticated bool
	Flash           string
}

type BlogUserPair struct {
	Blog *models.Blog
	User *models.User
}

var functions = template.FuncMap{
	"zipBlogsToUsers": zipBlogsToUsers,
	"humanDate":       humanDate,
}

func zipBlogsToUsers(
	blogs []*models.Blog,
	users []*models.User,
) []BlogUserPair {
	var results []BlogUserPair
	for i := range blogs {
		results = append(results, BlogUserPair{
			Blog: blogs[i],
			User: users[i],
		})
	}
	return results
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	base := "html/base.tmpl"
	nav := "html/partials/nav.tmpl"
	search := "html/partials/search.tmpl"
	miniBlog := "html/pages/mini-blog.tmpl"
	patterns := []string{
		base,
		nav,
		search,
		miniBlog,
	}
	for _, name := range []string{"home", "search", "myBlogs", "userBlogs", "blog"} {
		patterns = append(patterns, fmt.Sprintf("html/pages/%v.tmpl", name))
		templateSet, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = templateSet

	}

	patterns = []string{
		base,
		nav,
	}
	for _, name := range []string{"login", "signup", "about", "create", "update", "account", "changePassword"} {
		patterns = append(patterns, fmt.Sprintf("html/pages/%v.tmpl", name))
		templateSet, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = templateSet
	}

	return cache, nil
}
