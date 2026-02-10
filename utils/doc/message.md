{{range .Messages}}
### {{.Package}}.{{.Name}}
{{range .Comment}}{{.}}
{{end}}
**Fields:**

| Name | Type | Required | Tags | Description |
|---|---|---|---|---|
{{range .Fields}}| {{.Name}} | {{.Type}} | {{.Required}} | {{range $k, $v := .Tags}}{{$k}}:{{$v}} {{end}} | {{range .Comment}}{{.}} {{end}} |
{{end}}

{{end}}
