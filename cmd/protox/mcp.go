package main

import (
	"bytes"
	"os"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	mcpHelperOnce sync.Once
)

func generateMCPHelpCode(filename string, goPackage string) error {
	var e error
	mcpHelperOnce.Do(func() {
		file, err := os.Create(filename)
		if err != nil {
			e = err
			return
		}
		defer file.Close()

		genWriteln(file, "package %s", goPackage)
		genWriteln(file, `
import (
	context "context"

	connect "connectrpc.com/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func generateToolFunc[In, Out any](f func(context.Context, *In) (*Out, error), interceptors ...connect.Interceptor) func(context.Context, *mcp.CallToolRequest, *In) (*mcp.CallToolResult, *Out, error) {
	h := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		out, err := f(ctx, req.Any().(*In))
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(out), nil
	}
	
	for i := len(interceptors) - 1; i >= 0; i-- {
		h = interceptors[i].WrapUnary(h)
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input *In) (*mcp.CallToolResult, *Out, error) {
		r := connect.NewRequest(input)
		for k := range req.Extra.Header {
			r.Header().Set(k, req.Extra.Header.Get(k))
		}
		out, err := h(ctx, r)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{}, out.Any().(*Out), nil
	}
}`)

	})
	return e
}

func generateMCPServer(f *protogen.File, m *protogen.Service, value any) error {
	dir := *outDir + "connect"
	goPackage := dir

	filename := path.Join(*outDir, dir, path.Base(f.GeneratedFilenamePrefix)+".mcp.go")
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return err
	}

	if err := generateMCPHelpCode(path.Join(*outDir, dir, "mcp_help.go"), goPackage); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	px := &pxFile{
		goPackage: path.Base(goPackage),
		buf:       &bytes.Buffer{},
		File:      file,
	}
	defer px.Close()

	opt := value.(*protox.MCPOption)
	if !opt.Enable {
		return nil
	}

	px.Import("github.com/modelcontextprotocol/go-sdk/mcp")
	px.Import("github.com/google/jsonschema-go/jsonschema")
	px.Import(`emptypb "google.golang.org/protobuf/types/known/emptypb"`)
	px.Import(`connect "connectrpc.com/connect"`)

	pkgName := string(f.GoPackageName)

	px.Writeln("var _ emptypb.Empty")

	px.Writeln("")
	px.Writeln("func Add%sMcpTools(s *mcp.Server, svc %sHandler, interceptors ...connect.Interceptor) {", m.GoName, m.GoName)
	for _, method := range m.Methods {
		generateMCPMethod(px, method, pkgName)
	}
	px.Writeln("}")
	return nil
}

func generateMCPMethod(px *pxFile, method *protogen.Method, pkgName string) {
	px.WritelnIndent(1, "mcp.AddTool(s,")
	px.WritelnIndent(2, "&mcp.Tool{")
	px.WritelnIndent(3, "Name: %q,", method.GoName)

	if desc := trimComment(method.Comments); desc != "" {
		px.WritelnIndent(3, "Description: %q,", desc)
	}
	px.WritelnIndent(3, "InputSchema: jsonschema.Schema{")
	px.WritelnIndent(4, "Type: \"object\",")
	generateSchema(px, method.Input, 4)
	px.WritelnIndent(3, "},") // end of InputSchema

	px.WritelnIndent(3, "OutputSchema: jsonschema.Schema{")
	px.WritelnIndent(4, "Type: \"object\",")
	generateSchema(px, method.Output, 4)
	px.WritelnIndent(3, "},") // end of OutputSchema

	px.WritelnIndent(2, "},") // end of mcp.Tool

	px.WritelnIndent(2, "generateToolFunc(svc.%s, interceptors...),", method.GoName)
	px.WritelnIndent(1, ")") // end of mcp.AddTool
}

func generateSchema(px *pxFile, message *protogen.Message, indent int) {
	required := []string{}
	px.WritelnIndent(indent, "Properties: map[string]*jsonschema.Schema{")
	for _, field := range message.Fields {
		jsonName := string(field.Desc.Name())
		gotags := make(map[string]string)
		opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Tag) {
			gotags = parseGoTags(proto.GetExtension(opt, protox.E_Tag).(string))
			if parts := strings.Split(gotags["json"], ","); parts[0] != "" {
				jsonName = parts[0]
			}
		}
		if jsonName == "-" {
			continue
		}
		if slices.Contains(strings.Split(gotags["vd"], ","), "required") {
			required = append(required, jsonName)
		}

		px.WritelnIndent(indent+1, "%q: {", jsonName)
		if field.Desc.IsList() {
			px.WritelnIndent(indent+2, "Type: \"array\",")
			px.WritelnIndent(indent+2, "Items: &jsonschema.Schema{")
			generateFieldType(px, field, indent+3)
			px.WritelnIndent(indent+2, "},")
		} else {
			generateFieldType(px, field, indent+2)
		}

		if desc := trimComment(field.Comments); desc != "" {
			px.WritelnIndent(indent+2, "Description: %q,", desc)
		}
		px.WritelnIndent(indent+1, "},")
	}
	px.WritelnIndent(indent, "},")

	// generate Required field
	if len(required) > 0 {
		px.WritelnIndent(indent, "Required: []string{")
		for _, v := range required {
			px.WritelnIndent(indent+1, "%q,", v)
		}
		px.WritelnIndent(indent, "},")
	}
}

func generateFieldType(px *pxFile, field *protogen.Field, indent int) {

	kind := field.Desc.Kind()
	switch kind {
	case protoreflect.MessageKind:
		// special cases
		switch field.Desc.Message().FullName() {
		case "protox.Timestamp":
			px.WritelnIndent(indent, "Type: \"integer\",")
		case "google.protobuf.Empty":
			px.WritelnIndent(indent, "Type: \"object\",")
		default:
			px.WritelnIndent(indent, "Type: \"object\",")
			generateSchema(px, field.Message, indent)
		}

	case protoreflect.EnumKind:
		px.WritelnIndent(indent, "Type: \"string\",")
		px.WritelnIndent(indent, "Enum: []any{")
		for _, v := range field.Enum.Values {
			px.WritelnIndent(indent+1, "%q,", v.Desc.Name())
		}
		px.WritelnIndent(indent, "},")
	default:
		jsonType := "string"
		switch kind {
		case protoreflect.BoolKind:
			jsonType = "boolean"
		case protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind,
			protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind,
			protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind,
			protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
			jsonType = "integer"
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			jsonType = "number"
		case protoreflect.StringKind:
			jsonType = "string"
		case protoreflect.BytesKind, protoreflect.GroupKind:
			panic("group and bytes kind is not supported for mcp server")
		}
		px.WritelnIndent(indent, "Type: %q,", jsonType)
	}
}

func trimComment(comments protogen.CommentSet) string {
	desc := strings.TrimLeft(comments.Leading.String(), "/")
	for _, c := range comments.LeadingDetached {
		desc += " " + strings.TrimLeft(c.String(), "/")
	}
	desc += " " + strings.TrimLeft(comments.Trailing.String(), "/")
	desc = strings.ReplaceAll(desc, "\"", "\\\"")
	desc = strings.ReplaceAll(desc, "\n", " ")
	return strings.TrimSpace(desc)
}

func parseGoTags(tag string) map[string]string {
	result := make(map[string]string)
	if tag == "" {
		return result
	}

	for tag != "" {
		// Skip leading spaces
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Find the colon separator
		colonIdx := strings.Index(tag, ":")
		if colonIdx < 0 {
			break
		}

		key := tag[:colonIdx]
		tag = tag[colonIdx+1:]

		// Expect opening quote
		if len(tag) == 0 || tag[0] != '"' {
			break
		}

		// Find closing quote, handling escaped quotes
		tag = tag[1:]
		valueStart := 0
		j := 0
		for j < len(tag) {
			if tag[j] == '\\' && j+1 < len(tag) {
				j += 2
			} else if tag[j] == '"' {
				break
			} else {
				j++
			}
		}

		if j >= len(tag) {
			break
		}

		value := tag[valueStart:j]
		result[key] = value

		// Move past the closing quote
		tag = tag[j+1:]
	}

	return result
}
