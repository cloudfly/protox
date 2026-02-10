package main

import (
	"slices"
	"strings"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func generateGoJSONMarshaler(px *pxFile, m *protogen.Message) error {
	jsonNames := make(map[string]JSONOption)
	defined := false
	for _, field := range m.Fields {
		opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Json) {
			value := proto.GetExtension(opt, protox.E_Json)
			if value.(string) == "" {
				jsonNames[field.GoName] = JSONOption{Name: field.GoName}
			} else {
				defined = true
				items := strings.Split(value.(string), ",")
				jsonNames[field.GoName] = JSONOption{
					Name:      items[0],
					ReadOnly:  slices.Contains(items[1:], "readonly"),
					WriteOnly: slices.Contains(items[1:], "writeonly"),
				}
			}
		} else {
			jsonNames[field.GoName] = JSONOption{Name: field.GoName}
		}
	}

	if !defined {
		// no any json option defined, skip
		return nil
	}

	px.Import("encoding/json")

	if err := generateGoJSONMarshal(px, m, jsonNames); err != nil {
		return err
	}
	if err := generateGoJSONUnmarshal(px, m, jsonNames); err != nil {
		return err
	}

	return nil
}

func generateGoJSONMarshal(px *pxFile, m *protogen.Message, jsonNames map[string]JSONOption) error {
	px.Writeln("")
	px.Writeln("func (x %s) JSON() ([]byte, error) {", m.GoIdent.GoName)
	defer px.Writeln("}")
	px.WritelnIndent(1, "data := map[string]any{")
	for _, field := range m.Fields {
		if c := field.GoName[0]; c >= 'A' && c <= 'Z' {
			// only handle exported Go Field
			opt, ok := jsonNames[field.GoName]
			if !ok {
				continue
			}
			if opt.Name == "-" || opt.WriteOnly {
				// if field is writeonly, represent it should not be marshaled into JSON
				continue
			}
			px.WritelnIndent(2, `"%s": x.%s,`, opt.Name, field.GoName)
		}
	}
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return json.Marshal(data)")
	return nil
}

func generateGoJSONUnmarshal(f *pxFile, m *protogen.Message, jsonNames map[string]JSONOption) error {
	f.Writeln("")
	f.Writeln("func (x *%s) FromJSON(content []byte) (error) {", m.GoIdent.GoName)
	defer f.Writeln("}")
	f.WritelnIndent(1, "data := map[string]any{")
	for _, field := range m.Fields {
		if c := field.GoName[0]; c >= 'A' && c <= 'Z' {
			// only handle exported Go Field
			opt, ok := jsonNames[field.GoName]
			if !ok {
				continue
			}
			if opt.Name == "-" || opt.ReadOnly {
				// if field is readonly, represent it should not be unmarshaled from JSON
				continue
			}
			f.WritelnIndent(2, `"%s": &x.%s,`, opt.Name, field.GoName)
		}
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, "return json.Unmarshal(content, &data)")
	return nil

}
