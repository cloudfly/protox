package doc

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

//go:embed index.html
var html string

//go:embed service.md
var serviceTemplate string

//go:embed message.md
var messageTemplate string

var (
	serviceTmpl *template.Template
	messageTmpl *template.Template
)

func init() {
	var err error
	serviceTmpl, err = template.New("doc").Parse(serviceTemplate)
	if err != nil {
		panic(err)
	}
	messageTmpl, err = template.New("doc").Parse(messageTemplate)
	if err != nil {
		panic(err)
	}
}

func errorHandler(err error) http.HandlerFunc {
	msg := err.Error()
	n := fmt.Sprintf("%d", len(msg))
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", n)
		w.Write([]byte(err.Error()))
	}
}

func responseError(w http.ResponseWriter, err error) {
	msg := err.Error()
	n := fmt.Sprintf("%d", len(msg))
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", n)
	w.Write([]byte(err.Error()))
}

func MarkdownHandler(messages []*Message, services []*Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("message") != "" {
			serveMessageRequest(w, r, messages)
		} else {
			serveServiceRequest(w, r, services)
		}
	}
}

func serveServiceRequest(w http.ResponseWriter, r *http.Request, services []*Service) {
	filteredServices := make([]*Service, 0, len(services))
	search := r.URL.Query().Get("method")
	if search != "" {
		for _, svc := range services {
			data := make([]Method, 0, len(services))
			for _, method := range svc.Methods {
				if strings.Contains(method.Name, search) {
					data = append(data, method)
				}
			}
			if len(data) > 0 {
				svcCopy := &Service{
					Name:    svc.Name,
					Comment: svc.Comment,
					Package: svc.Package,
					Methods: data,
				}
				filteredServices = append(filteredServices, svcCopy)
			}
		}
	} else {
		filteredServices = services
	}

	buf := &bytes.Buffer{}
	err := serviceTmpl.Execute(buf, map[string]any{
		"Services": filteredServices,
	})
	if err != nil {
		responseError(w, err)
		return
	}
	body := buf.Bytes()
	bodyLength := fmt.Sprintf("%d", len(body))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", bodyLength)
	w.Write(body)

}

func serveMessageRequest(w http.ResponseWriter, r *http.Request, messages []*Message) {
	data := make([]*Message, 0, len(messages))
	search := r.URL.Query().Get("message")

	if search != "" {
		for _, msg := range messages {
			if strings.Contains(msg.Name, search) {
				data = append(data, msg)
			}
		}
	} else {
		data = messages
	}

	buf := &bytes.Buffer{}
	err := messageTmpl.Execute(buf, map[string]any{
		"Messages": data,
	})
	if err != nil {
		responseError(w, err)
		return
	}
	body := buf.Bytes()
	bodyLength := fmt.Sprintf("%d", len(body))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", bodyLength)
	w.Write(body)
}

func HtmlHandler(messages []*Message, services []*Service) http.HandlerFunc {
	tmpl, err := template.New("doc").Parse(html)
	if err != nil {
		return errorHandler(err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, map[string]any{
		"Messages": messages,
		"Services": services,
	})
	if err != nil {
		return errorHandler(err)
	}
	body := buf.Bytes()
	bodyLength := fmt.Sprintf("%d", len(body))
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", bodyLength)
		w.Write(body)
	}
}

// Service represents information about an interface.
type Service struct {
	Name    string
	Comment []string
	Package string
	Methods []Method
}

// Method represents a method definition in an interface.
type Method struct {
	Name    string
	Comment []string
	Input   *Message // Input parameter type information
	Output  *Message // Return value type information
}

// Message represents information about a data type.
type Message struct {
	Package string
	Comment []string
	Name    string
	Fields  []Field
}

// Field represents a field in a struct.
type Field struct {
	Name     string
	Comment  []string
	Type     string
	Required bool
	Tags     map[string]string
}
