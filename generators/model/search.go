package model

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
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

// SearchNamespaceData stores namespace info for search template
type SearchNamespaceData struct {
	GeneratorVersion string
	Package          string

	HasImports bool
	Imports    []string

	GoPGVer string

	Entities []SearchEntityData
}

// PackSearchNamespace packs mfd namespace to template data
func PackSearchNamespace(namespaces []*mfd.Namespace, options Options) (SearchNamespaceData, error) {
	imports := mfd.NewSet()

	var models []SearchEntityData
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			mdl, err := PackSearchEntity(*entity, options)
			if err != nil {
				return SearchNamespaceData{}, fmt.Errorf("error packing search entity: %w", err)
			}
			if len(mdl.Columns) == 0 {
				continue
			}

			for _, imp := range mdl.Imports {
				imports.Add(imp)
			}

			models = append(models, mdl)
		}
	}

	goPGVer := ""
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return SearchNamespaceData{
		GeneratorVersion: mfd.Version,
		Package:          options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		GoPGVer: goPGVer,

		Entities: models,
	}, nil
}

// SearchEntityData stores entity info for template
type SearchEntityData struct {
	// using model template as base because search depends on it
	EntityData

	Columns []SearchAttributeData
	Imports []string
}

// PackSearchEntity packs mfd entity to template data
func PackSearchEntity(entity mfd.Entity, options Options) (SearchEntityData, error) {
	imports := util.NewSet()

	columns := make([]SearchAttributeData, 0, len(entity.Attributes))
	// adding search
	for _, attribute := range entity.Attributes {
		if attribute.IsArray || attribute.IsJSON() || attribute.IsMap() {
			continue
		}
		// adding simple search for every column
		column := PackSearchAttribute(entity, *attribute, options)
		columns = append(columns, column)
		if column.Import != "" {
			imports.Add(column.Import)
		}
	}

	// adding search from search section
	for _, search := range entity.Searches {
		column, err := CustomSearchAttribute(entity, *search, options)
		if err != nil {
			return SearchEntityData{}, fmt.Errorf("error generating search attribute: %w", err)
		}
		columns = append(columns, column)
	}

	return SearchEntityData{
		// base template entity
		EntityData: PackEntity(entity, options),

		Columns: columns,
		Imports: imports.Elements(),
	}, nil
}

// SearchAttributeData stores attribute info for template
type SearchAttributeData struct {
	// using model template as base because search depends on it
	AttributeData

	UseCustomRender bool
	CustomRender    template.HTML
}

// PackSearchAttribute packs mfd attribute to template data
func PackSearchAttribute(entity mfd.Entity, attribute mfd.Attribute, options Options) SearchAttributeData {
	column := PackAttribute(entity, attribute, options)

	column.GoType = mfd.MakeSearchType(column.GoType, mfd.SearchEquals)

	return SearchAttributeData{
		// base template entity
		AttributeData: column,
	}
}

// CustomSearchAttribute applies custom search filters by attributes
func CustomSearchAttribute(entity mfd.Entity, search mfd.Search, options Options) (SearchAttributeData, error) {
	// use default templateColumn as base
	templateColumn := PackSearchAttribute(entity, *search.Attribute, options)
	templateColumn.Name = search.Name
	if search.Attribute.IsJSON() {
		if search.GoType == "" {
			search.GoType = model.TypeInterface
		}
		templateColumn.GoType = mfd.MakeSearchType(search.GoType, search.SearchType)
	} else {
		templateColumn.GoType = mfd.MakeSearchType(templateColumn.GoType, search.SearchType)
	}

	filterType, found := mfd.FilterTypeBySearchType[search.SearchType]
	if !found {
		return SearchAttributeData{}, fmt.Errorf("unknown search type: %s", search.SearchType)
	}

	templateColumn.IsArray = filterType.IsArray

	// rendering custom template
	var buffer bytes.Buffer
	// should not fail
	_ = templates.Execute(&buffer, filterData{
		Table:        entity.Name,
		ShortVarName: mfd.ShortVarName(entity.Name),
		Column:       template.HTML(columnRef(entity, search)),
		Value:        templateColumn.Name,
		SearchType:   filterType.Name,
		Exclude:      filterType.ExcludeString(),
		NoPointer:    templateColumn.IsArray,
	})

	templateColumn.CustomRender = template.HTML(buffer.String())
	templateColumn.UseCustomRender = true

	return templateColumn, nil
}

func columnRef(entity mfd.Entity, search mfd.Search) string {
	if strings.Contains(search.AttrName, ".") {
		parts := strings.SplitN(search.AttrName, ".", 2)
		return fmt.Sprintf(`"%s.%s"`, util.Underscore(parts[0]), search.Attribute.DBName)
	}

	if mfd.IsJSON(search.AttrName) {
		parts := strings.Split(search.AttrName, mfd.JSONFieldSep)
		return strconv.Quote(search.Attribute.DBName + mfd.JSONFieldSep + strings.Join(parts[1:], mfd.JSONFieldSep))
	}

	return fmt.Sprintf("Columns.%s.%s", entity.Name, search.AttrName)
}
