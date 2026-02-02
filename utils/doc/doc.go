package doc

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
)

//go:embed index.html
var html string

//go:embed index.md
var markdown string

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

func MarkdownHandler(messages []*Message, services []*Service) http.HandlerFunc {
	tmpl, err := template.New("doc").Parse(markdown)
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
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", bodyLength)
		w.Write(body)
	}
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
