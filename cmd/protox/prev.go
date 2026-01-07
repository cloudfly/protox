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
				if err := handleMessageInherit(
					f, msg,
					proto.GetExtension(msg.Desc.Options(), protox.E_Inherit).(*protox.InheritOption),
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleMessageInherit(f *protogen.File, m *protogen.Message, value *protox.InheritOption) error {
	if value == nil {
		return nil
	}

	ok := false
	for _, msg := range f.Messages {
		if string(msg.Desc.FullName().Name()) == value.Message {
			for _, field := range msg.Fields {
				if !slices.Contains(value.Omit, string(field.Desc.Name())) {
					m.Fields = append(m.Fields, field)
				}
			}
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("inherit message %s not found", value.Message)
	}

	return nil
}
