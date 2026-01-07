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

	parentMsg := findMessage(gen, opt.Message, f.Proto.GetPackage())
	if parentMsg == nil {
		return fmt.Errorf("inherit message %s not found", opt.Message)
	}

	px.Writeln("")
	px.Writeln(fmt.Sprintf("func (x *%s) From%s(parent *%s) {", m.GoIdent.GoName, parentName, parentName))
	px.WritelnIndent(1, "if x == nil || parent == nil {")
	px.WritelnIndent(2, "return")
	px.WritelnIndent(1, "}")

	for _, f := range parentMsg.Fields {
		if slices.Contains(opt.Omit, string(f.Desc.Name())) {
			continue
		}
		px.WritelnIndent(1, fmt.Sprintf("x.%s = parent.%s", f.GoName, f.GoName))
	}

	px.WritelnIndent(1, `return`)
	px.Writeln("}\n")

	px.Writeln("")
	px.Writeln(fmt.Sprintf("func (x *%s) To%s() *%s {", m.GoIdent.GoName, parentName, parentName))
	px.WritelnIndent(1, "if x == nil {")
	px.WritelnIndent(2, "return &%s{}", parentName)
	px.WritelnIndent(1, "}")

	px.WritelnIndent(1, "target := &%s{}", parentName)

	for _, f := range parentMsg.Fields {
		if slices.Contains(opt.Omit, string(f.Desc.Name())) {
			continue
		}
		px.WritelnIndent(1, fmt.Sprintf("target.%s = x.%s", f.GoName, f.GoName))
	}

	px.WritelnIndent(1, `return target`)
	px.Writeln("}\n")

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
