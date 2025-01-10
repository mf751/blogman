package main

import (
	"html/template"

	"github.com/mf751/blogman/ui"
)

type templateData struct{}

var functions = template.FuncMap{}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	base := "html/base.tmpl"
	nav := "html/partials/nav.tmpl"
	search := "html/pages/search.tmpl"
	patterns := []string{
		base,
		nav,
		search,
		"html/pages/home.tmpl",
	}
	templateSet, err := template.New("home").Funcs(functions).ParseFS(ui.Files, patterns...)
	if err != nil {
		return nil, err
	}

	cache["home"] = templateSet

	return cache, nil
}
