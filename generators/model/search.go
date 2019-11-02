package model

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

var templates *template.Template

func init() {
	var err error
	templates, err = template.New("filters").Parse(filters)
	if err != nil {
		panic(err)
	}
}

// filter data passed to custom filters template
type filterData struct {
	Table        string
	ShortVarName string
	Column       template.HTML
	Value        string
	SearchType   string
	Exclude      string
	NoPointer    bool
}

// this code used for creating objects to render search templates

// TemplatePackage stores package info
type SearchTemplatePackage struct {
	Package string

	HasImports bool
	Imports    []string

	Entities []SearchTemplateEntity
}

// NewSearchTemplatePackage creates a package for template
func NewSearchTemplatePackage(namespaces mfd.Namespaces, options Options) SearchTemplatePackage {
	imports := mfd.NewSet()

	var models []SearchTemplateEntity
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			mdl := NewSearchTemplateEntity(*entity, options)
			if len(mdl.Columns) == 0 {
				continue
			}

			for _, imp := range mdl.Imports {
				imports.Add(imp)
			}

			models = append(models, mdl)
		}
	}

	return SearchTemplatePackage{
		Package: options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}
}

// SearchTemplateEntity stores struct info
type SearchTemplateEntity struct {
	// using model template as base because search depends on it
	TemplateEntity

	Columns []SearchTemplateColumn
	Imports []string
}

// NewSearchTemplateEntity creates an entity for template
func NewSearchTemplateEntity(entity mfd.Entity, options Options) SearchTemplateEntity {
	imports := util.NewSet()

	var columns []SearchTemplateColumn

	// adding search
	for _, attribute := range entity.Attributes {
		if attribute.IsArray || attribute.IsJSON() || attribute.IsMap() {
			continue
		}
		// adding simple search for every column
		column := NewSearchTemplateColumn(entity, *attribute, options)
		columns = append(columns, column)
		if column.Import != "" {
			imports.Add(column.Import)
		}
	}

	// adding search from search section
	for _, search := range entity.Searches {
		column := NewCustomTemplateColumn(entity, *search, options)
		columns = append(columns, column)
	}

	return SearchTemplateEntity{
		// base template entity
		TemplateEntity: NewTemplateEntity(entity, options),

		Columns: columns,
		Imports: imports.Elements(),
	}
}

// SearchTemplateColumn stores column info
type SearchTemplateColumn struct {
	// using model template as base because search depends on it
	TemplateColumn

	UseCustomRender bool
	CustomRender    template.HTML
}

// NewSearchTemplateColumn creates a column for template
func NewSearchTemplateColumn(entity mfd.Entity, attribute mfd.Attribute, options Options) SearchTemplateColumn {
	column := NewTemplateColumn(entity, attribute, options)

	// making pointer for search types
	column.GoType = fmt.Sprintf("*%s", column.Type)

	return SearchTemplateColumn{
		// base template entity
		TemplateColumn: column,
	}
}

// NewCustomTemplateColumn creates custom search column
func NewCustomTemplateColumn(entity mfd.Entity, search mfd.Search, options Options) SearchTemplateColumn {
	// use default templateColumn as base
	templateColumn := NewSearchTemplateColumn(entity, *search.Attribute, options)
	templateColumn.Name = search.Name

	// if need to change type (array searches)
	if typ := mfd.MakeSearchType(templateColumn.Type, search.SearchType); typ != templateColumn.Type {
		templateColumn.GoType = typ
	}

	// TODO Refactor
	var filterType, exclude string
	switch search.SearchType {
	case mfd.SearchEquals:
		filterType, exclude = "SearchTypeEquals", "false"
	case mfd.SearchArray:
		filterType, exclude = "SearchTypeArray", "false"
		templateColumn.IsArray = true
	case mfd.SearchG:
		filterType, exclude = "SearchTypeGreater", "false"
	case mfd.SearchGE:
		filterType, exclude = "SearchTypeGE", "false"
	case mfd.SearchL:
		filterType, exclude = "SearchTypeLess", "false"
	case mfd.SearchLE:
		filterType, exclude = "SearchTypeLE", "false"
	case mfd.SearchILike:
		filterType, exclude = "SearchTypeILike", "false"
	case mfd.SearchLike:
		filterType, exclude = "SearchTypeLike", "false"
	case mfd.SearchLeftILike:
		filterType, exclude = "SearchTypeLILike", "false"
	case mfd.SearchLeftLike:
		filterType, exclude = "SearchTypeLLike", "false"
	case mfd.SearchRightILike:
		filterType, exclude = "SearchTypeRILike", "false"
	case mfd.SearchRightLike:
		filterType, exclude = "SearchTypeRLike", "false"
	case mfd.SearchNotArray:
		filterType, exclude = "SearchTypeArray", "true"
		templateColumn.IsArray = true
	case mfd.SearchNotEquals:
		filterType, exclude = "SearchTypeEquals", "true"
	case mfd.SearchNull:
		filterType, exclude = "SearchTypeNull", "false"
	case mfd.SearchNotNull:
		filterType, exclude = "SearchTypeNull", "true"
	}

	// rendering custom template
	var buffer bytes.Buffer
	// should not fail
	templates.Execute(&buffer, filterData{
		Table:        entity.Name,
		ShortVarName: mfd.ShortVarName(entity.Name),
		Column:       template.HTML(columnRef(entity, search)),
		Value:        templateColumn.Name,
		SearchType:   filterType,
		Exclude:      exclude,
		NoPointer:    templateColumn.IsArray,
	})

	templateColumn.CustomRender = template.HTML(buffer.String())
	templateColumn.UseCustomRender = true

	return templateColumn
}

func columnRef(entity mfd.Entity, search mfd.Search) string {
	if strings.Index(search.AttrName, ".") != -1 {
		parts := strings.SplitN(search.AttrName, ".", 2)
		return fmt.Sprintf(`"%s.%s"`, util.Underscore(parts[0]), search.Attribute.DBName)
	}

	return fmt.Sprintf("Columns.%s.%s", entity.Name, search.AttrName)
}
