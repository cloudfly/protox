{{range .Services}}
# Service: {{.Package}}.{{.Name}}
{{range .Comment}}{{.}}
{{end}}

## Methods
{{range .Methods}}
### {{.Name}}

{{range .Comment}}{{.}}
{{end -}}
- **Input:** [{{.Input.Package}}.{{.Input.Name}}]("?message={{.Input.Package}}.{{.Input.Name}}")
- **Output:** [{{.Input.Package}}.{{.Output.Name}}]("?message={{.Input.Package}}.{{.Output.Name}}")
{{end}}
{{end}}
