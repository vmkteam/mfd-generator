package model

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"

	"github.com/vmkteam/mfd-generator/mfd"
)

var templates *template.Template

var (
	searchExcluded = "true"
	searchIncluded = "false"
)

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

// NamespaceData stores namespace info for template
type SearchNamespaceData struct {
	GeneratorVersion string
	Package          string

	HasImports bool
	Imports    []string

	GoPGVer string

	Entities []SearchEntityData
}

// PackSearchNamespace packs mfd namespace to template data
func PackSearchNamespace(namespaces []*mfd.Namespace, options Options) SearchNamespaceData {
	imports := mfd.NewSet()

	var models []SearchEntityData
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			mdl := PackSearchEntity(*entity, options)
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
	}
}

// SearchEntityData stores entity info for template
type SearchEntityData struct {
	// using model template as base because search depends on it
	EntityData

	Columns []SearchAttributeData
	Imports []string
}

// PackSearchEntity packs mfd entity to template data
func PackSearchEntity(entity mfd.Entity, options Options) SearchEntityData {
	imports := util.NewSet()

	var columns []SearchAttributeData

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
		column := CustomSearchAttribute(entity, *search, options)
		columns = append(columns, column)
	}

	return SearchEntityData{
		// base template entity
		EntityData: PackEntity(entity, options),

		Columns: columns,
		Imports: imports.Elements(),
	}
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

// PackSearchAttribute packs mfd attribute to search template data
func CustomSearchAttribute(entity mfd.Entity, search mfd.Search, options Options) SearchAttributeData {
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
	// TODO Refactor
	var filterType, exclude string
	switch search.SearchType {
	case mfd.SearchEquals:
		filterType, exclude = "SearchTypeEquals", searchIncluded
	case mfd.SearchArray:
		filterType, exclude = "SearchTypeArray", searchIncluded
		templateColumn.IsArray = true
	case mfd.SearchG:
		filterType, exclude = "SearchTypeGreater", searchIncluded
	case mfd.SearchGE:
		filterType, exclude = "SearchTypeGE", searchIncluded
	case mfd.SearchL:
		filterType, exclude = "SearchTypeLess", searchIncluded
	case mfd.SearchLE:
		filterType, exclude = "SearchTypeLE", searchIncluded
	case mfd.SearchILike:
		filterType, exclude = "SearchTypeILike", searchIncluded
	case mfd.SearchLike:
		filterType, exclude = "SearchTypeLike", searchIncluded
	case mfd.SearchLeftILike:
		filterType, exclude = "SearchTypeLILike", searchIncluded
	case mfd.SearchLeftLike:
		filterType, exclude = "SearchTypeLLike", searchIncluded
	case mfd.SearchRightILike:
		filterType, exclude = "SearchTypeRILike", searchIncluded
	case mfd.SearchRightLike:
		filterType, exclude = "SearchTypeRLike", searchIncluded
	case mfd.SearchNotArray:
		filterType, exclude = "SearchTypeArray", searchExcluded
		templateColumn.IsArray = true
	case mfd.SearchNotEquals:
		filterType, exclude = "SearchTypeEquals", searchExcluded
	case mfd.SearchNull:
		filterType, exclude = "SearchTypeNull", searchIncluded
	case mfd.SearchNotNull:
		filterType, exclude = "SearchTypeNull", searchExcluded
	case mfd.SearchTypeArrayContains:
		filterType, exclude = "SearchTypeArrayContains", searchIncluded
		templateColumn.IsArray = false
	case mfd.SearchTypeArrayNotContains:
		filterType, exclude = "SearchTypeArrayContains", searchExcluded
		templateColumn.IsArray = false
	case mfd.SearchTypeArrayContained:
		filterType, exclude = "SearchTypeArrayContained", searchIncluded
		templateColumn.IsArray = true
	case mfd.SearchTypeArrayIntersect:
		filterType, exclude = "SearchTypeArrayIntersect", searchIncluded
		templateColumn.IsArray = true
	case mfd.SearchTypeJsonbPath:
		filterType, exclude = "SearchTypeJsonbPath", searchIncluded
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
	if strings.Contains(search.AttrName, ".") {
		parts := strings.SplitN(search.AttrName, ".", 2)
		return fmt.Sprintf(`"%s.%s"`, util.Underscore(parts[0]), search.Attribute.DBName)
	}

	if mfd.IsJson(search.AttrName) {
		parts := strings.Split(search.AttrName, mfd.JsonFieldSep)
		return strconv.Quote(search.Attribute.DBName + mfd.JsonFieldSep + strings.Join(parts[1:], mfd.JsonFieldSep))
	}

	return fmt.Sprintf("Columns.%s.%s", entity.Name, search.AttrName)
}
