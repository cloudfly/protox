package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func prevProcess(gen *protogen.Plugin) error {
	req := gen.Request

	for i, f := range gen.Files {
		pf := f.Proto
		pkg := pf.GetPackage()
		for _, msg := range pf.MessageType {
			if proto.HasExtension(msg.GetOptions(), protox.E_Inherit) {
				opt := proto.GetExtension(msg.GetOptions(), protox.E_Inherit).(*protox.InheritOption)
				parent := findProtoMessage(req, msg.GetName(), pkg)
				if parent == nil {
					return fmt.Errorf("inherit message %s not found", msg.GetName())
				}
				if err := handleMessageInherit(msg, parent, opt); err != nil {
					return err
				}
			}
		}

		filename := req.FileToGenerate[i]
		if err := os.Rename(filename, filename+".bak"); err != nil {
			return err
		}

	}

	return nil
}

func findProtoMessage(req *pluginpb.CodeGeneratorRequest, name string, fromPkg string) *descriptorpb.DescriptorProto {
	for _, pf := range req.ProtoFile {
		pkg := pf.GetPackage()
		for _, msg := range pf.MessageType {
			if pkg == fromPkg {
				// inherit message in same file
				if msg.GetName() == name {
					return msg
				}
			} else {
				// inherit a message from other package
				if pkg+"."+msg.GetName() == name {
					return msg
				}
			}
		}
	}
	return nil
}

func handleMessageInherit(msg, parent *descriptorpb.DescriptorProto, opt *protox.InheritOption) error {
	for _, field := range parent.Field {
		if !slices.Contains(opt.Omit, string(field.GetName())) {
			msg.Field = append(msg.Field, field)
		}
	}
	return nil
}
