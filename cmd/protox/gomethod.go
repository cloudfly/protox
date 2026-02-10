package main

import (
	"fmt"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
)

func generateGoMethods(f *pxFile, m *protogen.Message, value any) error {
	methods := value.([]*protox.GoMethodOption)
	for _, method := range methods {
		switch {
		case method.ReturnString != nil:
			f.Writeln("")
			f.WriteString(fmt.Sprintf(`func (%s) %s() string { return "%s" }`, m.GoIdent.GoName, method.Name, method.GetReturnString()))
			f.Writeln("")
		case method.ReturnInt64 != nil:
			f.Writeln("")
			f.WriteString(fmt.Sprintf(`func (%s) %s() int64 { return %d }`, m.GoIdent.GoName, method.Name, method.GetReturnInt64()))
			f.Writeln("")
		case method.ReturnBool != nil:
			f.Writeln("")
			f.WriteString(fmt.Sprintf(`func (%s) %s() bool { return %t }`, m.GoIdent.GoName, method.Name, method.GetReturnBool()))
			f.Writeln("")
		}
	}
	return nil
}
