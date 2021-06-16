package vt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/vmkteam/mfd-generator/mfd"
)

// updateServer updates namespaces constants of vt package.
// It is a improved version of PrintServer.
func updateServer(namespaces []*mfd.VTNamespace, options Options) error {
	pack, err := PackServerNamespaces(namespaces, options)
	if err != nil {
		return fmt.Errorf("packing data error: %w", err)
	}

	filename := options.Output
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse server dir: %w", err)
	}

	entities := make(map[string]EntityData)
	foundEntities := make(map[string]bool)
	for i, model := range pack.Entities {
		foundEntities[model.VarName] = false
		entities[model.VarName] = pack.Entities[i]
	}

	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			var changed bool // do not touch file if there is nothing to change

			newFile := astutil.Apply(file, func(cursor *astutil.Cursor) bool {
				if cursor.Name() != "Decls" {
					return true // skip
				}

				// ensure top-level declaration is GenDecl
				decl, ok := cursor.Node().(*ast.GenDecl)
				if !ok {
					return true
				}
				if decl.Doc == nil {
					return true // skip
				}

				var isNamespaces bool
				for _, cmnt := range decl.Doc.List {
					if cmnt == nil {
						continue
					}
					if cmnt.Text == "//mfd:rpc-namespaces" {
						isNamespaces = true
						break
					}
				}
				if !isNamespaces {
					return true // skip
				}

				// search for currently represented entities
				for _, spec := range decl.Specs {
					valSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					// we can work only with 1 const declaration per line
					//identName := valSpec.Names[0].Name
					if len(valSpec.Values) == 0 {
						// iota declaration
						continue
					}

					valueContent := strings.Trim(valSpec.Values[0].(*ast.BasicLit).Value, `"`)

					// ensure that we expect this entity:
					// there may be other custom rpc namespaces that we dont want to touch
					if _, ok := foundEntities[valueContent]; ok {
						foundEntities[valueContent] = true
					}
				}

				for name, found := range foundEntities {
					if found {
						continue
					}
					ent := entities[name]

					newConst := &ast.ValueSpec{
						Names: []*ast.Ident{{
							Name: "NS" + ent.Name,
						}},
						Values: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"` + ent.VarName + `"`}},
					}

					decl.Specs = append(decl.Specs, newConst)
					changed = true
				}

				return true
			}, nil)

			if changed {
				var b bytes.Buffer
				err = format.Node(&b, fset, newFile)
				if err != nil {
					return fmt.Errorf("format file: %w", err)
				}

				if err := ioutil.WriteFile(fileName, b.Bytes(), os.ModePerm); err != nil {
					return fmt.Errorf("write updated file: %w", err)
				}
			}
		}
	}
	return nil
}
