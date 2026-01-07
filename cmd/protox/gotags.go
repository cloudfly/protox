package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/cloudfly/protox"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	rInject = regexp.MustCompile("`.+`$")
	rTags   = regexp.MustCompile(`[\w_]+:"[^"]+"`)
)

type textArea struct {
	Start        int
	End          int
	CurrentTag   string
	InjectTag    string
	CommentStart int
	CommentEnd   int
}

func generateGoTag(px *pxFile, f *protogen.File, generatedContent []byte) error {
	data := make(map[string]string)
	for _, m := range f.Messages {
		for _, field := range m.Fields {
			if opt, _ := field.Desc.Options().(*descriptorpb.FieldOptions); opt != nil {
				if proto.HasExtension(opt, protox.E_Gotag) {
					data[field.GoIdent.GoName] = proto.GetExtension(opt, protox.E_Gotag).(string)
				}
			}
		}
	}
	filename := path.Join(*outDir, f.GeneratedFilenamePrefix+".pb.go")

	areas, err := parseGoFile(filename, generatedContent, data)
	if err != nil {
		return err
	}

	for i := range areas {
		area := areas[len(areas)-i-1]
		log.Info().Msgf("inject custom tag %q to expression %q", area.InjectTag, string(generatedContent[area.Start-1:area.End-1]))
		generatedContent = injectTag(generatedContent, area)
	}
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return err
	}
	if err = os.WriteFile(filename, generatedContent, 0o644); err != nil {
		return err
	}
	return nil
}

func parseGoFile(inputPath string, src []byte, tags map[string]string) (areas []textArea, err error) {
	log.Info().Msgf("parsing file %q for injecting tag with content(len=%d)", inputPath, len(src))
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputPath, src, parser.ParseComments)
	if err != nil {
		return
	}

	for _, decl := range f.Decls {
		// check if is generic declaration
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var typeSpec *ast.TypeSpec
		for _, spec := range genDecl.Specs {
			if ts, tsOK := spec.(*ast.TypeSpec); tsOK {
				typeSpec = ts
				break
			}
		}

		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}

		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		for _, field := range structDecl.Fields.List {
			tag := tags[typeSpec.Name.Name+"_"+field.Names[0].Name]
			if tag == "" {
				continue
			}

			currentTag := field.Tag.Value
			area := textArea{
				Start:      int(field.Pos()),
				End:        int(field.End()),
				CurrentTag: currentTag[1 : len(currentTag)-1],
				InjectTag:  tag,
			}
			areas = append(areas, area)
		}
	}
	log.Info().Msgf("parsed file %q, number of fields to inject custom tags: %d", inputPath, len(areas))
	return
}

type tagItem struct {
	key   string
	value string
}

type tagItems []tagItem

func (ti tagItems) format() string {
	tags := []string{}
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func (ti tagItems) override(nti tagItems) tagItems {
	overrided := []tagItem{}
	for i := range ti {
		dup := -1
		for j := range nti {
			if ti[i].key == nti[j].key {
				dup = j
				break
			}
		}
		if dup == -1 {
			overrided = append(overrided, ti[i])
		} else {
			overrided = append(overrided, nti[dup])
			nti = append(nti[:dup], nti[dup+1:]...)
		}
	}
	return append(overrided, nti...)
}

func newTagItems(tag string) tagItems {
	items := []tagItem{}
	splitted := rTags.FindAllString(tag, -1)

	for _, t := range splitted {
		sepPos := strings.Index(t, ":")
		items = append(items, tagItem{
			key:   t[:sepPos],
			value: t[sepPos+1:],
		})
	}
	return items
}

func injectTag(contents []byte, area textArea) (injected []byte) {
	expr := make([]byte, area.End-area.Start)
	copy(expr, contents[area.Start-1:area.End-1])
	cti := newTagItems(area.CurrentTag)
	iti := newTagItems(area.InjectTag)
	ti := cti.override(iti)
	expr = rInject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))

	injected = append(injected, contents[:area.Start-1]...)
	injected = append(injected, expr...)
	injected = append(injected, contents[area.End-1:]...)

	return
}
