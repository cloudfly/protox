package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func generateApiProto(w io.Writer, f *protogen.File, m *protogen.Message, options *protox.ApiOption) error {
	generateFilterMessage(w, m)
	generatePatchMessage(w, m)
	if !options.NoGet {
		if err := generateGetRequest(w, m); err != nil {
			return err
		}
	}
	if !options.NoCreate {
		if err := generateCreateRequest(w, m); err != nil {
			return err
		}
	}
	if !options.NoUpdate {
		if err := generateUpdateRequest(w, m); err != nil {
			return err
		}
	}
	if !options.NoSelect {
		if err := generateSelectRequest(w, f, m); err != nil {
			return err
		}
	}
	if !options.NoDelete {
		if err := generateDeleteRequest(w, m); err != nil {
			return err
		}
	}

	if err := generateService(w, f, m, options); err != nil {
		return err
	}

	return nil
}

func getFieldClass(field *protogen.Field, prefix string) ([]string, bool) {
	var classes []string
	if opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions); opt != nil {
		if proto.HasExtension(opt, protox.E_Class) {
			classes = strings.Fields(proto.GetExtension(opt, protox.E_Class).(string))
		}
	}

	k, ok := 0, false
	for _, item := range classes {
		if item == prefix || strings.HasPrefix(item, prefix+":") {
			ok = true
			if len(item) > len(prefix)+1 {
				classes[k] = item[len(prefix)+1:]
				k++
			}
		}
	}

	classes = classes[:k]
	options := make([]string, 0, len(classes))
	for _, class := range classes {
		options = append(options, strings.Split(class, ",")...)
	}
	return options, ok
}

func generateCreateRequest(w io.Writer, m *protogen.Message) error {
	genWriteln(w, "message Create%sRequest {", m.Desc.Name())
	defer genWritelnln(w, "}")

	for _, field := range m.Fields {
		classes, ok := getFieldClass(field, "create")
		if !ok {
			continue
		}
		genWritelnIndent(w, 1, "%s %s = %d;",
			fieldProtoType(field.Desc, !slices.Contains(classes, "required")),
			field.Desc.Name(),
			field.Desc.Index(),
		)
	}

	return nil
}

func generateFilterMessage(w io.Writer, m *protogen.Message) {
	genWriteln(w, "message %sFilter {", m.Desc.Name())
	defer genWritelnln(w, "}")
	generateFilterMessageFields(w, m)
}

func generateFilterMessageFields(w io.Writer, m *protogen.Message) {
	for _, field := range m.Fields {
		classes, ok := getFieldClass(field, "filter")
		if !ok {
			continue
		}
		genWritelnIndent(w, 1, "%s %s = %d;",
			fieldProtoType(field.Desc, !slices.Contains(classes, "required")),
			field.Desc.Name(),
			field.Desc.Index(),
		)
	}
}

func generatePatchMessage(w io.Writer, m *protogen.Message) {
	genWriteln(w, "message %sPatch {", m.Desc.Name())
	defer genWritelnln(w, "}")

	for _, field := range m.Fields {
		classes, ok := getFieldClass(field, "patch")
		if !ok {
			continue
		}
		_ = classes
		genWritelnIndent(w, 1, "%s %s = %d;",
			fieldProtoType(field.Desc, true),
			field.Desc.Name(),
			field.Desc.Index(),
		)
	}
}

func generateUpdateRequest(w io.Writer, m *protogen.Message) error {
	genWriteln(w, "message Update%sRequest {", m.Desc.Name())
	defer genWritelnln(w, "}")

	genWritelnIndent(w, 1, "%sFilter filter = 1;", m.Desc.Name())
	genWritelnIndent(w, 1, "%sPatch patch = 2;", m.Desc.Name())

	return nil
}

func generateSelectRequest(w io.Writer, f *protogen.File, m *protogen.Message) error {
	genWriteln(w, "message List%sRequest {", m.Desc.Name())
	generateFilterMessageFields(w, m)
	genWritelnln(w, "}")

	genWriteln(w, "message List%sResponse {", m.Desc.Name())
	genWritelnIndent(w, 1, "repeated %s.%s items = 1;", f.Proto.GetPackage(), m.Desc.Name())
	genWritelnIndent(w, 1, "int64 total = 2;")
	genWritelnln(w, "}")

	return nil
}

func generateGetRequest(w io.Writer, m *protogen.Message) error {
	genWriteln(w, "message Get%sRequest {", m.Desc.Name())
	defer genWritelnln(w, "}")

	genWritelnIndent(w, 1, "string id = 1;")
	return nil
}

func generateDeleteRequest(w io.Writer, m *protogen.Message) error {
	genWriteln(w, "message Delete%sRequest {", m.Desc.Name())
	defer genWritelnln(w, "}")

	genWritelnIndent(w, 1, "string id = 1;")
	return nil
}

func generateService(w io.Writer, f *protogen.File, m *protogen.Message, options *protox.ApiOption) error {
	messageName := m.Desc.Name()
	serviceName := options.GetServiceName()
	if serviceName == "" {
		serviceName = fmt.Sprintf("%sGeneratedApiService", messageName)
	}

	genWriteln(w, "service %s {", serviceName)
	defer genWriteln(w, "}")

	if !options.NoCreate {
		genWritelnIndent(w, 1, "rpc Create%s(Create%sRequest) returns (%s.%s);", messageName, messageName, f.Proto.GetPackage(), messageName)
	}
	if !options.NoUpdate {
		genWritelnIndent(w, 1, "rpc Update%s(Update%sRequest) returns (google.protobuf.Empty);", messageName, messageName)
	}
	if !options.NoSelect {
		genWritelnIndent(w, 1, "rpc List%s(List%sRequest) returns (List%sResponse);", messageName, messageName, messageName)
	}
	if !options.NoGet {
		genWritelnIndent(w, 1, "rpc Get%s(Get%sRequest) returns (%s.%s);", messageName, messageName, f.Proto.GetPackage(), messageName)
	}
	if !options.NoDelete {
		genWritelnIndent(w, 1, "rpc Delete%s(Delete%sRequest) returns (google.protobuf.Empty);", messageName, messageName)
	}

	genWritelnIndent(w, 1, "option (protox.docx) = {")
	genWritelnIndent(w, 2, "enable: true,")
	genWritelnIndent(w, 2, "targetMessageName: \"%s\",", m.Desc.Name())
	genWritelnIndent(w, 2, "targetMessageGoName: \"%s\",", m.GoIdent.GoName)
	genWritelnIndent(w, 2, "targetMessageLocation: \"%s\",", m.Location.SourceFile)
	genWritelnIndent(w, 2, "targetMessageGoImportPath: %s,", m.GoIdent.GoImportPath)
	genWritelnIndent(w, 1, "};")

	return nil
}

func fieldProtoType(field protoreflect.FieldDescriptor, optional bool) string {
	prefix := ""
	if field.IsList() {
		prefix = "repeated "
	} else if field.IsMap() {
		prefix = fmt.Sprintf("map<%s, %s> ", fieldProtoType(field.MapKey(), false), fieldProtoType(field.MapValue(), false))
	} else if optional {
		prefix = "optional "
	}

	if msg := field.Message(); msg != nil {
		return prefix + string(msg.FullName())
	}
	return prefix + field.Kind().String()
}
