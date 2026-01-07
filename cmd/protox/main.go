package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"unsafe"

	"github.com/cloudfly/protox"
	"github.com/k0kubun/pp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
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
	filenames := []string{}
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		plugin = gen
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for p, f := range gen.FilesByPath {
			if f.GoImportPath == "github.com/cloudfly/protox" {
				// ignore the protox.proto
				continue
			}
			if strings.HasPrefix(string(f.GoImportPath), "google.golang.org") {
				continue
			}

			if err := handleFile(p, f); err != nil {
				return err
			}
			filenames = append(filenames, p)
		}
		return nil
	})
}

func handleFile(filename string, f *protogen.File) error {

	log.Info().Str("file", filename).Msg("Handling proto file ...")

	px, err := newPxFile(filename, f)
	if err != nil {
		return err
	}
	defer px.Close()

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
		if proto.HasExtension(opt, protox.E_Gomethod) {
			if err := generateGoMethods(px, m, proto.GetExtension(opt, protox.E_Gomethod)); err != nil {
				return err
			}
		}
		if proto.HasExtension(opt, protox.E_Gosql) {
			if err := generateSerializer(px, m, proto.GetExtension(opt, protox.E_Gosql)); err != nil {
				return err
			}
		}
	}

	if err := generateGoJSONMarshaler(px, m); err != nil {
		return err
	}

	if err := generateOrmxFieldOption(px, m); err != nil {
		return err
	}

	return nil
}

func handleField(f *protogen.File, m *protogen.Message, field *protogen.Field) error {
	if field == nil {
		return nil
	}
	opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions)
	if opt == nil {
		return nil
	}
	return nil
}

func generateGoError(f *pxFile, e *protogen.Enum) error {
	errMsgs := make(map[string]string)
	defined := false
	for _, ev := range e.Values {
		opt, _ := ev.Desc.Options().(*descriptorpb.EnumValueOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Goerror) {
			defined = true
			value := proto.GetExtension(opt, protox.E_Goerror)
			errMsgs[ev.GoIdent.GoName] = value.(string)
		} else {
			errMsgs[ev.GoIdent.GoName] = ev.GoIdent.GoName
		}
	}

	if !defined {
		// no any json option defined, skip
		return nil
	}

	f.Writeln("")
	f.Writeln("func (x %s) Error() string {", e.GoIdent.GoName)
	defer f.Writeln("}")
	f.WritelnIndent(1, "switch x {")
	for k, v := range errMsgs {
		f.WritelnIndent(2, "case %s: ", k)
		f.WritelnIndent(3, "return %q", v)
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, "return fmt.Sprintf(\"unknown %s %%d\", x)", e.GoIdent.GoName)
	return nil
}

func generateOrmxFieldOption(f *pxFile, m *protogen.Message) error {
	ormxOptions := make(map[string]*protox.OrmxOption)
	defined := false
	for _, field := range m.Fields {
		opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Ormx) {
			value := proto.GetExtension(opt, protox.E_Ormx)
			if value == nil {
				ormxOptions[field.GoName] = &protox.OrmxOption{Column: string(field.Desc.Name())}
			} else {
				defined = true
				oo := value.(*protox.OrmxOption)
				if oo.Column == "" {
					oo.Column = string(field.Desc.Name()) // use proto field name by default
				}
				ormxOptions[field.GoName] = oo
			}
		} else {
			ormxOptions[field.GoName] = &protox.OrmxOption{Column: string(field.Desc.Name())}
		}
	}

	if !defined {
		// no any ormx option defined, skip
		return nil
	}

	return generateOrmxColumnOptionMethod(f, m, ormxOptions)
}

func generateOrmxColumnOptionMethod(f *pxFile, m *protogen.Message, opts map[string]*protox.OrmxOption) error {
	f.Writeln("")
	f.Writeln("func (x %s) OrmxColumnOption(fieldName string) string {", m.GoIdent.GoName)

	optstrs := make(map[string]string)
	for name, opt := range opts {
		s := &strings.Builder{}
		s.WriteString(opt.GetColumn()) // column must can not be empty
		if opt.Insert != nil {
			s.WriteString(fmt.Sprintf(",insert:%t", *opt.Insert))
		}
		if opt.Select != nil {
			s.WriteString(fmt.Sprintf(",select:%t", *opt.Select))
		}
		if opt.Update != nil {
			s.WriteString(fmt.Sprintf(",update:%t", *opt.Update))
		}
		if opt.Incr != nil {
			s.WriteString(fmt.Sprintf(",incr:%d", opt.GetIncr()))
		}
		if opt.Desr != nil {
			s.WriteString(fmt.Sprintf(",decr:%d", opt.GetDesr()))
		}
		if op := opt.GetOperate(); op != "" {
			s.WriteString(fmt.Sprintf(",op:%s", op))
		}
		if t := opt.GetType(); t != "" {
			s.WriteString(fmt.Sprintf("type:%s", t))
		}
		optstrs[name] = s.String()
	}

	f.WritelnIndent(1, "switch fieldName {")
	for name, optstr := range optstrs {
		f.WritelnIndent(2, `case "%s": `, name)
		f.WritelnIndent(3, "return %q", optstr)
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, `return ""`)
	f.Writeln("}")

	f.Writeln("")
	f.Writeln("func (x %s) OrmxColumn(fieldName string) string {", m.GoIdent.GoName)
	f.WritelnIndent(1, "switch fieldName {")
	for name, opt := range opts {
		f.WritelnIndent(2, `case "%s": `, name)
		f.WritelnIndent(3, "return %q", opt.Column)
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, `return ""`)

	f.Writeln("}")

	return nil
}

func generateGoJSONMarshaler(f *pxFile, m *protogen.Message) error {
	jsonNames := make(map[string]JSONOption)
	defined := false
	for _, field := range m.Fields {
		opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions)
		if opt != nil && proto.HasExtension(opt, protox.E_Gojson) {
			value := proto.GetExtension(opt, protox.E_Gojson)
			if value.(string) == "" {
				jsonNames[field.GoName] = JSONOption{Name: field.GoName}
			} else {
				defined = true
				items := strings.Split(value.(string), ",")
				jsonNames[field.GoName] = JSONOption{
					Name:      items[0],
					ReadOnly:  slices.Contains(items[1:], "readonly"),
					WriteOnly: slices.Contains(items[1:], "writeonly"),
				}
			}
		} else {
			jsonNames[field.GoName] = JSONOption{Name: field.GoName}
		}
	}

	if !defined {
		// no any json option defined, skip
		return nil
	}

	if err := generateGoJSONMarshal(f, m, jsonNames); err != nil {
		return err
	}
	if err := generateGoJSONUnmarshal(f, m, jsonNames); err != nil {
		return err
	}

	return nil
}

func generateGoJSONMarshal(f *pxFile, m *protogen.Message, jsonNames map[string]JSONOption) error {
	f.Writeln("")
	f.Writeln("func (x %s) JSON() ([]byte, error) {", m.GoIdent.GoName)
	defer f.Writeln("}")
	f.WritelnIndent(1, "data := map[string]any{")
	for _, field := range m.Fields {
		if c := field.GoName[0]; c >= 'A' && c <= 'Z' {
			// only handle exported Go Field
			opt, ok := jsonNames[field.GoName]
			if !ok {
				continue
			}
			if opt.Name == "-" || opt.WriteOnly {
				// if field is writeonly, represent it should not be marshaled into JSON
				continue
			}
			f.WritelnIndent(2, `"%s": x.%s,`, opt.Name, field.GoName)
		}
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, "return json.Marshal(data)")
	return nil
}

func generateGoJSONUnmarshal(f *pxFile, m *protogen.Message, jsonNames map[string]JSONOption) error {
	f.Writeln("")
	f.Writeln("func (x *%s) FromJSON(content []byte) (error) {", m.GoIdent.GoName)
	defer f.Writeln("}")
	f.WritelnIndent(1, "data := map[string]any{")
	for _, field := range m.Fields {
		if c := field.GoName[0]; c >= 'A' && c <= 'Z' {
			// only handle exported Go Field
			opt, ok := jsonNames[field.GoName]
			if !ok {
				continue
			}
			if opt.Name == "-" || opt.ReadOnly {
				// if field is readonly, represent it should not be unmarshaled from JSON
				continue
			}
			f.WritelnIndent(2, `"%s": &x.%s,`, opt.Name, field.GoName)
		}
	}
	f.WritelnIndent(1, "}")
	f.WritelnIndent(1, "return json.Unmarshal(content, &data)")
	return nil

}

func generateGoMethods(f *pxFile, m *protogen.Message, value any) error {
	genWriteln(f, "")
	def := value.(*protox.GoMethodOption)
	if def.Return == nil {
		f.WriteString(fmt.Sprintf(`func (*%s) %s() { }`, m.GoIdent.GoName, def.Name))
	} else {
		f.WriteString(fmt.Sprintf(`func (*%s) %s() string { return "%s" }`, m.GoIdent.GoName, def.Name, def.GetReturn()))
	}
	genWriteln(f, "")
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
