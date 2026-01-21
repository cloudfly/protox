package utils

import (
	_ "embed"
	"html/template"
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	services = map[string]*protogen.Service{}
)

// embed: index.html
var html string

func Register(svc *protogen.Service) {
	key := string(svc.Desc.FullName())
	services[key] = svc
}

func Handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(html))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, map[string]any{
		"Services": services,
	})
}
