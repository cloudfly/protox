package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
)

func generateGoDocxService(px *pxFile, f *protogen.File, s *protogen.Service, options *protox.ApiOption) error {
	if !options.Enable {
		return nil
	}

	targetMessageName, err := getTargetMessageFullname(options)
	if err != nil {
		return err
	}
	handlerName := fmt.Sprintf("%sDocxHandler", s.GoName)
	px.Writeln("")
	px.Writeln(fmt.Sprintf(`type %s struct {}`, handlerName))
	px.Writeln("")

	for _, method := range s.Methods {
		px.Import("context", "github.com/cloudfly/docx")
		px.Writeln("func (s *%s) %s(ctx context.Context, req *%s) (*%s, error) {", handlerName, method.GoName, fullGoTypeName(px, f, method.Input), fullGoTypeName(px, f, method.Output))
		switch {
		case strings.HasPrefix(method.GoName, "Create"):
			generateGoDocxCreateMethodCode(px, method, targetMessageName)
		case strings.HasPrefix(method.GoName, "Update"):
			generateGoDocxUpdateMethodCode(px)
		case strings.HasPrefix(method.GoName, "List"):
			generateGoDocxListMethodCode(px, f, method, targetMessageName)
		case strings.HasPrefix(method.GoName, "Get"):
			generateGoDocxGetMethodCode(px, targetMessageName)
		case strings.HasPrefix(method.GoName, "Delete"):
			generateGoDocxDeleteMethodCode(px, targetMessageName)
		}
		px.Writeln("}\n")
	}

	return generateGoServiceRegister(f, s)
}

func generateGoServiceRegister(f *protogen.File, s *protogen.Service) error {
	goPackageName := string(f.GoPackageName) + "doc"
	targetDir := path.Join(path.Dir(f.GeneratedFilenamePrefix), goPackageName)
	targetFile := path.Join(targetDir, path.Base(f.Proto.GetName()))
	px, err := newPxFile(targetFile, goPackageName)
	if err != nil {
		return err
	}
	defer px.Close()

	connectPackageName := string(f.GoPackageName) + "connect"
	px.Import(string(f.GoImportPath) + "/" + connectPackageName)
	px.Import(`docutil "github.com/cloudfly/protox/utils/doc"`)

	px.Writeln("func init() {")
	px.WritelnIndent(1, "docutil.Register((*%s.%sHandler)(nil))", connectPackageName, s.GoName)
	px.Writeln("}")
	return nil
}

func generateGoDocxCreateMethodCode(px *pxFile, method *protogen.Method, targetMessageName string) {
	px.Import("github.com/creasty/defaults")

	px.WritelnIndent(1, `data := %s{}`, targetMessageName)
	px.WritelnIndent(1, "if err := defaults.Set(&data); err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.Writeln("")

	for _, field := range method.Input.Fields {
		if !field.Desc.HasOptionalKeyword() {
			px.WritelnIndent(1, fmt.Sprintf("data.%s = req.%s", field.GoName, field.GoName))
		}
	}
	px.Writeln("")
	for _, field := range method.Input.Fields {
		if field.Desc.HasOptionalKeyword() {
			px.WritelnIndent(1, fmt.Sprintf("if req.%s != nil {", field.GoName))
			px.WritelnIndent(2, fmt.Sprintf("data.%s = *req.%s", field.GoName, field.GoName))
			px.WritelnIndent(1, "}")
		}
	}

	px.Writeln("")

	px.WritelnIndent(1, "result, err := docx.Insert(ctx, &data)")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.Writeln("")
	px.WritelnIndent(1, "return result, nil")
}
func generateGoDocxUpdateMethodCode(px *pxFile) {
	px.Import(`emptypb "google.golang.org/protobuf/types/known/emptypb"`)
	px.WritelnIndent(1, "if err := docx.UpdateWhere(ctx, req.Filter, req.Patch); err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &emptypb.Empty{}, nil")
}
func generateGoDocxListMethodCode(px *pxFile, f *protogen.File, method *protogen.Method, targetMessageName string) {
	px.WritelnIndent(1, "var data []*%s", targetMessageName)
	px.WritelnIndent(1, "if err := docx.SelectWhere(ctx, req, &data); err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &%s{", fullGoTypeName(px, f, method.Output))
	px.WritelnIndent(2, "Items: data,")
	px.WritelnIndent(1, "}, nil")
}
func generateGoDocxGetMethodCode(px *pxFile, targetMessageName string) {
	px.WritelnIndent(1, "var data %s", targetMessageName)
	px.WritelnIndent(1, "if err := docx.GetById(ctx, req.Id, &data); err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &data, nil")
}
func generateGoDocxDeleteMethodCode(px *pxFile, targetMessageName string) {
	px.Import(`emptypb "google.golang.org/protobuf/types/known/emptypb"`)
	px.WritelnIndent(1, "if err := docx.DeleteById(ctx, req.Id, docx.Collection(%s{})); err != nil {", targetMessageName)
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &emptypb.Empty{}, nil")
}

func getTargetMessageFullname(options *protox.ApiOption) (string, error) {
	f, ok := plugin.FilesByPath[options.GetTargetMessageLocation()]
	if !ok {
		return "", fmt.Errorf("target message location %s not found", options.TargetMessageLocation)
	}
	for _, msg := range f.Messages {
		if string(msg.Desc.Name()) == options.GetTargetMessageName() {
			return fmt.Sprintf("%s.%s", f.GoPackageName, msg.GoIdent.GoName), nil
		}
	}
	return "", fmt.Errorf("target message name %s not found", options.TargetMessageName)
}

func fullGoTypeName(px *pxFile, f *protogen.File, msg *protogen.Message) string {
	if f.Desc.Path() == msg.Location.SourceFile {
		return msg.GoIdent.GoName
	}
	outputFile := plugin.FilesByPath[msg.Location.SourceFile]
	if outputFile != nil {
		px.Import(fmt.Sprintf("%s %s", outputFile.GoPackageName, outputFile.GoImportPath))
		return fmt.Sprintf("%s.%s", outputFile.GoPackageName, msg.GoIdent.GoName)
	}
	return msg.GoIdent.GoName
}
