package main

import (
	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func generateGoError(px *pxFile, e *protogen.Enum) error {
	errMsgs := make(map[string]string)
	defined := false
	for _, ev := range e.Values {
		opt, _ := ev.Desc.Options().(*descriptorpb.EnumValueOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Error) {
			defined = true
			value := proto.GetExtension(opt, protox.E_Error)
			errMsgs[ev.GoIdent.GoName] = value.(string)
		} else {
			errMsgs[ev.GoIdent.GoName] = ev.GoIdent.GoName
		}
	}

	if !defined {
		// no any json option defined, skip
		return nil
	}

	px.Import("fmt")

	px.Writeln("")
	px.Writeln("func (x %s) Error() string {", e.GoIdent.GoName)
	defer px.Writeln("}")
	px.WritelnIndent(1, "switch x {")
	for k, v := range errMsgs {
		px.WritelnIndent(2, "case %s: ", k)
		px.WritelnIndent(3, "return %q", v)
	}
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return fmt.Sprintf(\"unknown %s %%d\", x)", e.GoIdent.GoName)
	return nil
}
