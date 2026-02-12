package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"unsafe"

	"github.com/cloudfly/protox"
	"github.com/k0kubun/pp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

var _ = pp.Println

func init() {
	log.Logger = log.Logger.With().Caller().Logger().Output(&zerolog.ConsoleWriter{Out: os.Stderr})
}

var (
	flags  = flag.NewFlagSet("protox", flag.ContinueOnError)
	outDir = flags.String("out", ".", "the output directory")
)

var (
	plugin *protogen.Plugin
)

func main() {
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		plugin = gen

		var (
			services = make(map[string][]*protogen.Service)
			messages = make(map[string][]*protogen.Message)
			packages = make(map[string]string)
		)

		for filename, f := range gen.FilesByPath {
			if shouldGenerate(f) {
				gendFile := gengo.GenerateFile(gen, f)
				// skip the file, so that this file will be ignored in Response and no related .pb.go file will be generated,
				// Therefore the generateGoTag() will generate .pb.go file instead.
				gendFile.Skip()

				if err := generateGos(filename, gendFile, f); err != nil {
					return err
				}
				dir := path.Dir(filename)
				packages[dir] = string(f.GoPackageName)
				services[dir] = append(services[dir], f.Services...)
				messages[dir] = append(messages[dir], f.Messages...)
			}
		}

		for dir, pkg := range packages {
			docFileName := path.Join(*outDir, dir, "protox.doc.go")
			if err := generateDocs(docFileName, pkg, services[dir], messages[dir]); err != nil {
				return err
			}
		}

		return nil
	})
}

func shouldGenerate(f *protogen.File) bool {
	if !f.Generate {
		return false
	}
	/*
		if f.GoImportPath == "github.com/cloudfly/protox" {
			// ignore the protox.proto
			return false
		}
	*/
	if strings.HasPrefix(string(f.GoImportPath), "google.golang.org") {
		return false
	}
	return true
}

func generateGos(filename string, gendFile *protogen.GeneratedFile, f *protogen.File) error {
	px, err := newPxFile(filename, string(f.GoPackageName))
	if err != nil {
		return err
	}
	defer px.Close()

	log.Info().Str("file", filename).Bool("generate", f.Generate).Msg("Generating go tags ...")
	content, err := gendFile.Content()
	if err != nil {
		return err
	}
	if err := generateGoTag(px, f, content); err != nil {
		return err
	}
	log.Info().Str("file", filename).Msg("Completed generating .pb.go file")

	if err := handleFile(px, f); err != nil {
		return err
	}
	log.Info().Str("file", filename).Msg("Completed handling .px.go file ...")
	return nil
}

func handleFile(px *pxFile, f *protogen.File) error {
	for _, message := range f.Messages {
		if err := handleMessage(px, f, message); err != nil {
			return err
		}
	}
	for _, enum := range f.Enums {
		if err := handleEnum(px, f, enum); err != nil {
			return err
		}
	}
	for _, service := range f.Services {
		if err := handleService(px, f, service); err != nil {
			return err
		}
	}

	return nil
}

func handleService(px *pxFile, f *protogen.File, s *protogen.Service) error {
	if opt, _ := s.Desc.Options().(*descriptorpb.ServiceOptions); opt != nil {
		if proto.HasExtension(opt, protox.E_Mcp) {
			if err := generateMCPServer(f, s, proto.GetExtension(opt, protox.E_Mcp)); err != nil {
				return err
			}
		}
	}
	return nil
}

func handleEnum(px *pxFile, f *protogen.File, e *protogen.Enum) error {
	if err := generateGoError(px, e); err != nil {
		return err
	}
	return nil
}

func handleMessage(px *pxFile, f *protogen.File, m *protogen.Message) error {
	if opt, _ := m.Desc.Options().(*descriptorpb.MessageOptions); opt != nil {
		if proto.HasExtension(opt, protox.E_Method) {
			if err := generateGoMethods(px, m, proto.GetExtension(opt, protox.E_Method)); err != nil {
				return err
			}
		}
		if proto.HasExtension(opt, protox.E_Sql) {
			if err := generateSerializer(px, m, proto.GetExtension(opt, protox.E_Sql)); err != nil {
				return err
			}
		}
		if proto.HasExtension(opt, protox.E_Orm) {
			if err := generateGoOrm(px, f, m, proto.GetExtension(opt, protox.E_Orm)); err != nil {
				return err
			}
		}
	}

	if err := generateGoJSONMarshaler(px, m); err != nil {
		return err
	}

	return nil
}

func generateSerializer(px *pxFile, m *protogen.Message, value any) error {
	spec := value.(*protox.SQLOption)
	if spec.Serializer == nil {
		return nil
	}
	switch strings.ToLower(*spec.Serializer) {
	case "json":
		return generateSerializerJSON(px, m)
	}
	return nil
}

func generateSerializerJSON(px *pxFile, m *protogen.Message) error {
	px.Import("database/sql/driver", "encoding/json")
	name := m.GoIdent.GoName
	genWriteln(px, "")
	genWriteln(px, "func (data *%s) Scan(src any) error {", name)
	genWritelnIndent(px, 1, `*data = %s{}
	if src == nil {
		return nil
	}
	var content []byte
	switch value := src.(type) {
	case string:
		content = []byte(value)
	case []byte:
		content = value
	default:
		return fmt.Errorf("can not convert %%#v into %s", src)
	}
	if len(content) == 0 {
		return nil
	}
	return json.Unmarshal(content, data)`, name, name)
	genWritelnln(px, "}")

	genWriteln(px, "func (data %s) Value() (driver.Value, error) {", name)
	genWritelnIndent(px, 1, `content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return string(content), nil`)
	genWritelnln(px, "}")

	return nil
}

func pretty(v ...interface{}) {
	pp.Fprintln(os.Stderr, v...)
}

type pxFile struct {
	started   bool
	imports   []string
	goPackage string
	buf       *bytes.Buffer
	*os.File
}

func newPxFile(filename string, goPackage string) (*pxFile, error) {
	filename = path.Join(*outDir, strings.ReplaceAll(filename, ".proto", ".px.go"))
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return nil, err
	}
	px, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &pxFile{
		goPackage: goPackage,
		buf:       &bytes.Buffer{},
		File:      px,
	}, nil
}

func (f *pxFile) Write(data []byte) (n int, err error) {
	if !f.started {
		if _, err = f.File.WriteString("package " + f.goPackage + "\n\n"); err != nil {
			return
		}

		f.started = true
	}

	return f.buf.Write(data)
}

func (f *pxFile) WriteString(s string) (n int, err error) {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return f.Write(b)
}

func (f *pxFile) Close() error {

	genImport(f.File, f.imports...)

	if _, err := io.Copy(f.File, f.buf); err != nil {
		return err
	}
	f.buf.Reset()
	f.imports = nil

	if err := f.File.Close(); err != nil {
		return err
	}
	if !f.started {
		return os.Remove(f.File.Name())
	}
	return nil
}

func (f *pxFile) Import(imports ...string) {
	for _, item := range imports {
		if !slices.Contains(f.imports, item) {
			f.imports = append(f.imports, item)
		}
	}
}

func (f *pxFile) Writeln(format string, args ...any) {
	genWriteln(f, format, args...)
}

func (f *pxFile) WritelnIndent(indent int, format string, args ...any) {
	genWritelnIndent(f, indent, format, args...)
}

type JSONOption struct {
	Name      string
	ReadOnly  bool
	WriteOnly bool
}
