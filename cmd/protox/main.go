package main

import (
	"flag"
	"os"
	"path"
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
	flags    = flag.NewFlagSet("protox", flag.ContinueOnError)
	outDir   = flags.String("out", ".", "the output directory")
	rootDir  = flags.String("root", "", "the root directory of protos")
	genProto = flags.String("gen-proto", "", "generate the proto files")
)

func main() {
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures

		for filename, f := range gen.FilesByPath {
			if shouldGenerate(f) {
				gendFile := gengo.GenerateFile(gen, f)
				// skip the file, so that this file will be ignored in Response and no related .pb.go file will be generated,
				// Therefore the generateGoTag() will generate .pb.go file instead.
				gendFile.Skip()

				var err error
				if *genProto != "" {
					log.Info().Str("root", *rootDir).Str("file", filename).Msg("Generating extend proto files...")
					err = generateProtos(f)
				} else {
					err = generateGos(filename, gendFile, f)
				}
				if err != nil {
					return err
				}
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

func generateProtos(f *protogen.File) error {
	for _, m := range f.Messages {
		if opt, _ := m.Desc.Options().(*descriptorpb.MessageOptions); opt != nil {
			if proto.HasExtension(opt, protox.E_Api) {
				if err := generateApiProto(f, m, proto.GetExtension(opt, protox.E_Api).(*protox.ApiOption)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateGos(filename string, gendFile *protogen.GeneratedFile, f *protogen.File) error {
	px, err := newPxFile(filename, f)
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
		/*
			if proto.HasExtension(opt, protox.E_Inherit) {
				if err := generateGoInherit(px, gen, f, m, proto.GetExtension(opt, protox.E_Inherit)); err != nil {
					return err
				}
			}
		*/
	}

	if err := generateGoJSONMarshaler(px, m); err != nil {
		return err
	}

	return nil
}

func generateSerializer(w *pxFile, m *protogen.Message, value any) error {
	spec := value.(*protox.SQLOption)
	if spec.Serializer == nil {
		return nil
	}
	switch strings.ToLower(*spec.Serializer) {
	case "json":
		return generateSerializerJSON(w, m)
	}
	return nil
}

func generateSerializerJSON(f *pxFile, m *protogen.Message) error {
	name := m.GoIdent.GoName
	genWriteln(f, "")
	genWriteln(f, "func (data *%s) Scan(src any) error {", name)
	genWritelnIndent(f, 1, `*data = %s{}
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
	genWritelnln(f, "}")

	genWriteln(f, "func (data %s) Value() (driver.Value, error) {", name)
	genWritelnIndent(f, 1, `content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return string(content), nil`)
	genWritelnln(f, "}")

	return nil
}

func pretty(v ...interface{}) {
	pp.Fprintln(os.Stderr, v...)
}

type pxFile struct {
	started   bool
	protoFile *protogen.File
	*os.File
}

func newPxFile(filename string, f *protogen.File) (*pxFile, error) {
	filename = path.Join(*outDir, strings.ReplaceAll(filename, ".proto", ".px.go"))
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return nil, err
	}
	px, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &pxFile{
		protoFile: f,
		File:      px,
	}, nil
}

func (f *pxFile) Write(data []byte) (n int, err error) {
	if !f.started {
		if _, err = f.File.WriteString("package " + string(f.protoFile.GoPackageName) + "\n\n"); err != nil {
			return
		}
		genImport(f.File, "database/sql/driver", "encoding/json", "fmt", "context", "errors")
		genWriteln(f.File, "var (")
		genWritelnIndent(f.File, 1, "_ = fmt.Printf")
		genWritelnIndent(f.File, 1, "_ = driver.Bool")
		genWritelnIndent(f.File, 1, "_ = json.Marshal")
		genWritelnIndent(f.File, 1, "_ = context.Background")
		genWritelnIndent(f.File, 1, "_ = errors.New")
		genWritelnln(f.File, ")")

		f.started = true
	}
	return f.File.Write(data)
}

func (f *pxFile) WriteString(s string) (n int, err error) {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return f.Write(b)
}

func (f *pxFile) Close() error {
	if err := f.File.Close(); err != nil {
		return err
	}
	if !f.started {
		return os.Remove(f.File.Name())
	}
	return nil
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
