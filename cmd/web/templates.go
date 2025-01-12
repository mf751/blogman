package main

import (
	"html/template"
	"time"

	"github.com/mf751/blogman/interanl/models"
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
		"html/pages/home.tmpl",
	}
	templateSet, err := template.New("home").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["home"] = templateSet

	patterns = []string{
		base,
		nav,
		search,
		"html/pages/blog.tmpl",
	}
	templateSet, err = template.New("blog").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["blog"] = templateSet

	patterns = []string{
		base,
		nav,
		"html/pages/login.tmpl",
	}
	templateSet, err = template.New("login").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["login"] = templateSet

	patterns = []string{
		base,
		nav,
		"html/pages/signup.tmpl",
	}
	templateSet, err = template.New("signup").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["signup"] = templateSet

	patterns = []string{
		base,
		nav,
		"html/pages/about.tmpl",
	}
	templateSet, err = template.New("about").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["about"] = templateSet

	patterns = []string{
		base,
		nav,
		"html/pages/createBlog.tmpl",
	}
	templateSet, err = template.New("create").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["create"] = templateSet

	patterns = []string{
		base,
		nav,
		search,
		miniBlog,
		"html/pages/myBlogs.tmpl",
	}
	templateSet, err = template.New("myBlogs").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}
	cache["myBlogs"] = templateSet

	return cache, nil
}
