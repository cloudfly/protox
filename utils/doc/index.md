# API Documentation

This document describes the Services and Messages defined in the API. It is intended to help understand the available functionality and data structures.

## Services

{{range .Services}}
### Service: {{.Package}}.{{.Name}}
{{range .Comment}}{{.}}
{{end}}

#### Methods

{{range .Methods}}
---
##### Method: {{.Name}}

{{range .Comment}}{{.}}
{{end}}

**Input:** `{{.Input.Package}}.{{.Input.Name}}`
**Output:** `{{.Input.Package}}.{{.Output.Name}}`

{{end}}
{{end}}

## Messages

{{range .Messages}}
### Message: {{.Package}}.{{.Name}}

{{range .Comment}}{{.}}
{{end}}

**Fields:**

| Name | Type | Required | Tags | Description |
|---|---|---|---|---|
{{range .Fields}}| {{.Name}} | {{.Type}} | {{.Required}} | {{range $k, $v := .Tags}}{{$k}}:{{$v}} {{end}} | {{range .Comment}}{{.}} {{end}} |
{{end}}

{{end}}
