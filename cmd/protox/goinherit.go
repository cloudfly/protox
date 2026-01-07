package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
)

func generateGoInherit(px *pxFile, gen *protogen.Plugin, f *protogen.File, m *protogen.Message, value any) error {
	opt := value.(*protox.InheritOption)
	parentName := opt.Message
	if i := strings.LastIndex(parentName, "."); i > 0 {
		parentName = parentName[i+1:]
	}

	px.Writeln("")
	px.Writeln(fmt.Sprintf("func (x *%s) From%s(parent *%s) *%s {", m.GoIdent.GoName, parentName, parentName, m.GoIdent.GoName))
	px.WritelnIndent(1, "if x == nil || parent == nil {")
	px.WritelnIndent(2, "return x")
	px.WritelnIndent(1, "}")

	parentMsg := findMessage(gen, opt.Message, f.Proto.GetPackage())

	for _, f := range parentMsg.Fields {
		if slices.Contains(opt.Omit, string(f.Desc.Name())) {
			continue
		}
		px.WritelnIndent(1, fmt.Sprintf("x.%s = parent.%s", f.GoName, f.GoName))
	}

	px.WritelnIndent(1, `return x`)
	px.Writeln("}")
	return nil
}

func findMessage(gen *protogen.Plugin, name string, fromPkg string) *protogen.Message {
	for _, f := range gen.Files {
		pkg := f.Proto.GetPackage()
		for _, msg := range f.Messages {
			if pkg == fromPkg {
				// inherit message in same file
				if string(msg.Desc.Name()) == name {
					return msg
				}
			} else {
				// inherit a message from other package
				if pkg+"."+string(msg.Desc.Name()) == name {
					return msg
				}
			}

		}
	}
	return nil
}
