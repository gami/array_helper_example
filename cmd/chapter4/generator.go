package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/format"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

//go:embed template
var templates embed.FS

type Generator struct {
	pkg    string
	files  []*ast.File
	typ    string
	fields []*Field
}

type Target struct {
	Pkg    string
	Typ    string
	Arr    string
	Fields []*Field
}

type Field struct {
	Name   string
	Arr    string
	Typ    string
	IsStar bool
}

func NewGenerator(dir string, typ string) (*Generator, error) {
	pkg, err := packageInfo(dir)

	if err != nil {
		return nil, err
	}

	fields := parseFields(typ, pkg.Syntax)

	return &Generator{
		pkg:    pkg.Name,
		files:  pkg.Syntax,
		typ:    typ,
		fields: fields,
	}, nil
}

func packageInfo(dir string) (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedSyntax,
		Tests: false,
	}, dir)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to load package dir=%v", dir)
	}

	if len(pkgs) != 1 {
		return nil, errors.Wrapf(err, "%d packages found", len(pkgs))
	}

	pkg := pkgs[0]

	return pkg, nil
}

func (g *Generator) Run() ([]byte, error) {
	w := &bytes.Buffer{}

	tmpl, err := template.ParseFS(templates, "template/*")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse templates")
	}

	t := &Target{
		Pkg:    g.pkg,
		Typ:    g.typ,
		Arr:    fmt.Sprintf("%ss", g.typ),
		Fields: g.fields,
	}

	err = tmpl.ExecuteTemplate(w, "function", t)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	fmted, err := format.Source(w.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "failed to format code")
	}

	return fmted, nil
}

func parseFields(typ string, files []*ast.File) []*Field {
	fields := make([]*Field, 0)

	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if ok && ts.Name.Name == typ {
				ast.Inspect(n, func(n ast.Node) bool {
					fld, ok := n.(*ast.Field)
					if !ok {
						// フィールド以外は無視する
						return true
					}

					var isStar bool
					var typ *ast.Ident

					star, ok := fld.Type.(*ast.StarExpr)
					if ok {
						// ポインタの場合
						isStar = true
						typ = star.X.(*ast.Ident)
					} else {
						// 値の場合
						isStar = false
						typ = fld.Type.(*ast.Ident)
					}

					fields = append(fields, &Field{
						Name:   fld.Names[0].Name,
						Typ:    typ.Name,
						IsStar: isStar,
					})

					return true
				})
				return false
			}
			return true
		})
	}

	return fields
}
