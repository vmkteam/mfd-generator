package model

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// GenerateParams packs json fields to params file
func GenerateParams(namespaces []*mfd.Namespace, output string, options Options) (bool, error) {
	paramsFile, err := ReadParamsFile(output, options.Package)
	if err != nil {
		return false, err
	}

	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			for _, attribute := range entity.Attributes {
				if attribute.IsJSON() {
					paramsFile.Add(entity.Name + attribute.Name)
				}
			}
		}
	}

	return paramsFile.Save(output)
}

type ParamsFile struct {
	set  *token.FileSet
	file *ast.File
}

// ReadParamsFile reads exiting params file
func ReadParamsFile(filename, pack string) (*ParamsFile, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		src := fmt.Sprintf("package %s", pack)
		if err := os.WriteFile(filename, []byte(src), 0644); err != nil {
			return nil, fmt.Errorf("write file, err=%w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("open file, err=%w", err)
	}

	set := token.NewFileSet()
	file, err := parser.ParseFile(set, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("open file, err=%w", err)
	}

	return &ParamsFile{set: set, file: file}, nil
}

// Has checks if params file has specific param
func (p *ParamsFile) Has(name string) bool {
	if len(p.file.Decls) == 0 {
		return false
	}

	for _, d := range p.file.Decls {
		typ, ok := d.(*ast.GenDecl)
		if !ok || typ.Tok != token.TYPE {
			continue
		}
		str, ok := typ.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		if str.Name.Name == name {
			return true
		}
	}

	return false
}

// Add adds new param to file
func (p *ParamsFile) Add(name string) bool {
	if p.Has(name) {
		return false
	}

	newDecl := &ast.GenDecl{
		TokPos: p.file.End(),
		Tok:    token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: name},
				Type: &ast.StructType{Fields: &ast.FieldList{}},
			},
		},
	}

	p.file.Decls = append(p.file.Decls, newDecl)
	return true
}

// Save saves params to filename
func (p *ParamsFile) Save(filename string) (bool, error) {
	var buffer bytes.Buffer
	if err := printer.Fprint(&buffer, p.set, p.file); err != nil {
		return false, fmt.Errorf("dump ast to file, err=%w", err)
	}

	return util.FmtAndSave(buffer.Bytes(), filename)
}
