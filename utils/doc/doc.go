package doc

import (
	_ "embed"
	"html/template"
	"net/http"
)

var (
	services = map[string]*ServiceInfo{}
)

// embed: index.html
var html string

func Register(svc interface{}) {
	ss, err := ParseService(svc)
	if err != nil {
		panic(err)
	}
	services[ss.Name] = ss
}

func Handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(html))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, map[string]any{
		"Services": services,
	})
}
