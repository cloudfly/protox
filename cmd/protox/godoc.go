package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func generateDocs(filename string, packageName string, services []*protogen.Service, messages []*protogen.Message) error {
	var err error
	docFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer docFile.Close()

	genWriteln(docFile, "package "+string(packageName)+"\n\n")
	genWriteln(docFile, "import \"github.com/cloudfly/protox/utils/doc\"\n\n")

	genWriteln(docFile, "var ProtoxMessages = []*doc.Message{")
	for _, msg := range messages {
		genWritelnIndent(docFile, 1, "{")
		if err := generateMessageDoc(docFile, msg, 2); err != nil {
			return err
		}
		genWritelnIndent(docFile, 1, "},")
	}
	genWriteln(docFile, "}\n")

	genWriteln(docFile, "var ProtoxServices = []*doc.Service{")
	defer genWriteln(docFile, "}")
	for _, svc := range services {
		if err := generateServiceDoc(docFile, svc, 1); err != nil {
			return err
		}
	}
	return nil
}

func generateServiceDoc(w io.Writer, svc *protogen.Service, tab int) error {
	genWritelnIndent(w, tab, "{")
	genWritelnIndent(w, tab+1, "Name: \"%s\",", string(svc.Desc.Name()))
	genWritelnIndent(w, tab+1, "Comment: []string{")
	generateStringList(w, svc.Comments.LeadingDetached, tab+2)
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab+1, "Package: \"%s\",", string(svc.Desc.ParentFile().Package()))
	genWritelnIndent(w, tab+1, "Methods: []doc.Method{")
	for _, method := range svc.Methods {
		if err := generateMethodDoc(w, method, tab+2); err != nil {
			return err
		}
	}
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab, "},")

	return nil
}

func generateMethodDoc(w io.Writer, method *protogen.Method, tab int) error {
	genWritelnIndent(w, tab, "{")
	genWritelnIndent(w, tab+1, "Name: \"%s\",", string(method.Desc.Name()))
	genWritelnIndent(w, tab+1, "Comment: []string{")
	generateStringList(w, method.Comments.LeadingDetached, tab+2)
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab+1, "Input: &doc.Message{")
	if err := generateMessageDoc(w, method.Input, tab+2); err != nil {
		return err
	}
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab+1, "Output: &doc.Message{")
	if err := generateMessageDoc(w, method.Output, tab+2); err != nil {
		return err
	}
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab, "},")
	return nil
}

func generateMessageDoc(w io.Writer, m *protogen.Message, tab int) error {
	genWritelnIndent(w, tab, "Name: \"%s\",", string(m.Desc.Name()))
	genWritelnIndent(w, tab, "Comment: []string{")
	generateStringList(w, m.Comments.LeadingDetached, tab+1)
	genWritelnIndent(w, tab, "},")
	genWritelnIndent(w, tab, "Package: \"%s\",", string(m.Desc.ParentFile().Package()))
	genWritelnIndent(w, tab, "Fields: []doc.Field{")
	for _, field := range m.Fields {
		if err := generateFieldDoc(w, field, tab+1); err != nil {
			return err
		}
	}
	genWritelnIndent(w, tab, "},")
	return nil
}

func generateFieldDoc(w io.Writer, f *protogen.Field, tab int) error {
	genWritelnIndent(w, tab, "{")
	genWritelnIndent(w, tab+1, "Name: \"%s\",", string(f.Desc.Name()))
	genWritelnIndent(w, tab+1, "Comment: []string{")
	generateStringList(w, f.Comments.LeadingDetached, tab+2)
	genWritelnIndent(w, tab+1, "},")
	genWritelnIndent(w, tab+1, "Type: %q,", typeName(f.Desc, true))
	genWritelnIndent(w, tab+1, "Required: %v,", !f.Desc.HasOptionalKeyword())
	if opt, _ := f.Desc.Options().(*descriptorpb.FieldOptions); opt != nil {
		if proto.HasExtension(opt, protox.E_Tag) {
			tag := proto.GetExtension(opt, protox.E_Tag).(string)
			if tag != "" {
				genWritelnIndent(w, tab+1, "Tags: map[string]string{")
				for k, v := range parseGoTags(tag) {
					genWritelnIndent(w, tab+2, fmt.Sprintf("%q: %s,", k, v))
				}
				genWritelnIndent(w, tab+1, "},")
			}
		}
	}
	genWritelnIndent(w, tab, "},")
	return nil
}

func generateStringList(w io.Writer, strs []protogen.Comments, indent int) {
	for _, s := range strs {
		genWritelnIndent(w, indent, "%q,", s)
	}
}

func typeName(f protoreflect.FieldDescriptor, recursive bool) string {
	if recursive {
		if f.IsList() {
			return fmt.Sprintf("[]%s", typeName(f, false))
		} else if f.IsMap() {
			return fmt.Sprintf("map[%s]%s", typeName(f.MapKey(), false), typeName(f.MapValue(), false))
		}
	}

	switch f.Kind() {
	case protoreflect.StringKind, protoreflect.EnumKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return "uint64"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "float64"
	case protoreflect.MessageKind:
		return string(f.Message().FullName())
	default:
		return f.Kind().String()
	}
}

func parseGoTags(tag string) map[string]string {
	tags := make(map[string]string)
	for _, item := range strings.Split(tag, " ") {
		kv := strings.Split(item, ":")
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}
