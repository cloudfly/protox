package main

import (
	"fmt"
	"slices"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func prevProcess(gen *protogen.Plugin) error {
	for _, f := range gen.Files {
		for _, msg := range f.Messages {
			if proto.HasExtension(msg.Desc.Options(), protox.E_Inherit) {
				opt := proto.GetExtension(msg.Desc.Options(), protox.E_Inherit).(*protox.InheritOption)
				parent := findMessage(gen, opt.Message, f.Proto.GetPackage())
				if parent == nil {
					return fmt.Errorf("inherit message %s not found", opt.Message)
				}
				if err := handleMessageInherit(msg, parent, opt); err != nil {
					return err
				}

			}
		}
	}
	return nil
}

func handleMessageInherit(msg *protogen.Message, parent *protogen.Message, opt *protox.InheritOption) error {
	for _, field := range parent.Fields {
		if !slices.Contains(opt.Omit, string(field.Desc.FullName())) {
			msg.Fields = append(msg.Fields, field)
		}
	}
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
