package main

import (
	"fmt"
	"strings"

	"github.com/cloudfly/go/types"
	"github.com/cloudfly/protox"
	"google.golang.org/protobuf/compiler/protogen"
)

func generateGoOrm(px *pxFile, f *protogen.File, m *protogen.Message, value any) error {
	opt := value.(*protox.OrmOption)
	if !opt.Enable {
		return nil
	}
	if opt.GetFilterMessage() == "" {
		opt.FilterMessage = types.Ptr(fmt.Sprintf("%sFilter", m.GoIdent.GoName))
	}
	if opt.GetPatchMessage() == "" {
		opt.PatchMessage = types.Ptr(fmt.Sprintf("%sPatch", m.GoIdent.GoName))
	}
	if opt.GetTable() == "" {
		opt.Table = types.Ptr(strings.ToLower(m.GoIdent.GoName))
	}

	px.Import("context", "github.com/cloudfly/ormx")

	// Table()
	px.Writeln("")
	px.Writeln("func (x *%s) Table() string {", m.GoIdent.GoName)
	px.WritelnIndent(1, "return %q", *opt.Table)
	px.Writeln("}\n")

	// Create
	px.Writeln("func Create%s(ctx context.Context, data *%s) (*%s, error) {", m.GoIdent.GoName, m.GoIdent.GoName, m.GoIdent.GoName)
	px.WritelnIndent(1, "if data == nil {")
	px.WritelnIndent(2, "return nil, nil")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "id, err := ormx.InsertOne(ctx, data, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "var data2 %s", m.GoIdent.GoName)
	px.WritelnIndent(1, "err = ormx.GetByID(ctx, &data2, id, ormx.CtxOption(ctx).FromMaster())")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &data2, nil")
	px.Writeln("}\n")

	// Update
	px.Writeln("func Update%s(ctx context.Context, filter *%s, patch *%s) error {", m.GoIdent.GoName, *opt.FilterMessage, *opt.PatchMessage)
	px.WritelnIndent(1, "if filter == nil || patch == nil {")
	px.WritelnIndent(2, "return nil")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "_, err := ormx.PatchWhere(ctx, patch, filter, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return nil")
	px.Writeln("}\n")

	// Delete
	px.Writeln("func Delete%s(ctx context.Context, filter *%s) error {", m.GoIdent.GoName, *opt.FilterMessage)
	px.WritelnIndent(1, "if filter == nil {")
	px.WritelnIndent(2, "return nil")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return ormx.DeleteWhere(ctx, filter, ormx.CtxOption(ctx))")
	px.Writeln("}\n")

	// Get
	px.Writeln("func Get%s(ctx context.Context, filter *%s) (*%s, error) {", m.GoIdent.GoName, *opt.FilterMessage, m.GoIdent.GoName)
	px.WritelnIndent(1, "var data %s", m.GoIdent.GoName)
	px.WritelnIndent(1, "err := ormx.GetWhere(ctx, &data, filter, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return nil, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return &data, nil")
	px.Writeln("}\n")

	// List
	px.Writeln("func List%s(ctx context.Context, filter *%s) ([]*%s, int64, error) {", m.GoIdent.GoName, *opt.FilterMessage, m.GoIdent.GoName)
	px.WritelnIndent(1, "data := []*%s{}", m.GoIdent.GoName)
	px.WritelnIndent(1, "err := ormx.SelectWhere(ctx, &data, filter, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil && !ormx.IsNotFound(err) {")
	px.WritelnIndent(2, "return nil, 0, err")
	px.WritelnIndent(1, "}")

	px.WritelnIndent(1, "count, err := ormx.Count(ctx, filter, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return nil, 0, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return data, count, nil")
	px.Writeln("}\n")

	// Count
	px.Writeln("func Count%s(ctx context.Context, filter *%s) (int64, error) {", m.GoIdent.GoName, *opt.FilterMessage)
	px.WritelnIndent(1, "count, err := ormx.Count(ctx, filter, ormx.CtxOption(ctx))")
	px.WritelnIndent(1, "if err != nil {")
	px.WritelnIndent(2, "return 0, err")
	px.WritelnIndent(1, "}")
	px.WritelnIndent(1, "return count, nil")
	px.Writeln("}")

	// Table() for Filter
	px.Writeln("")
	px.Writeln("func (x *%sFilter) Table() string {", m.GoIdent.GoName)
	px.WritelnIndent(1, "return %q", *opt.Table)
	px.Writeln("}\n")

	// Table() for Patch
	px.Writeln("")
	px.Writeln("func (x *%sPatch) Table() string {", m.GoIdent.GoName)
	px.WritelnIndent(1, "return %q", *opt.Table)
	px.Writeln("}\n")

	return nil
}
