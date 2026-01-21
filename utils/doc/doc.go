package doc

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
)

var (
	services = map[string]*ServiceInfo{}
)

//go:embed index.html
var html string

func Register(svc interface{}) {
	ss, err := ParseService(svc)
	if err != nil {
		panic(err)
	}
	services[ss.Name] = ss
}

func Handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("doc").Parse(html)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, map[string]any{
		"Services": services,
	})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))
	w.Write(buf.Bytes())
}
