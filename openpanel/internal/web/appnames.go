package web

import (
	"html/template"
	"strings"
)

// AppNames is every spelling of an app's name the shared CMS templates need, typed as JS so they also work as JS identifiers
type AppNames struct {
	Slug  template.JS // joomla
	Name  template.JS // OpenCart
	Title template.JS // Opencart
	Upper template.JS // OPENCART
	Camel template.JS // openCart
}

func NewAppNames(slug, name string) AppNames {
	return AppNames{
		Slug:  template.JS(slug),
		Name:  template.JS(name),
		Title: template.JS(strings.ToUpper(slug[:1]) + slug[1:]),
		Upper: template.JS(strings.ToUpper(slug)),
		Camel: template.JS(strings.ToLower(name[:1]) + name[1:]),
	}
}
