package repo

import (
	"fmt"
	"html/template"

	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// PKPair stores primary keys with type for template
type PKPair struct {
	Field string
	Arg   string
	Type  string
	Zero  template.HTML
}

// SortPair stores sort columns with direction for template
type SortPair struct {
	Field, Dir string
}

// NamespaceData stores namespace info for template
type NamespaceData struct {
	Package string

	Name         string
	ShortVarName string

	HasImports bool
	Imports    []string

	GoPGVer string

	Entities []EntityData
}

// PackNamespace packs mfd namespace to template data
func PackNamespace(namespace *mfd.Namespace, options Options) NamespaceData {
	imports := mfd.NewSet()
	entities := make([]EntityData, len(namespace.Entities))
	for i, entity := range namespace.Entities {
		packed := PackEntity(*entity, options)
		entities[i] = packed

		for _, imp := range packed.Imports {
			imports.Append(imp)
		}
	}

	name := util.CamelCased(util.Sanitize(namespace.Name))

	goPGVer := ""
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return NamespaceData{
		Package: options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Name:         name,
		ShortVarName: mfd.ShortVarName(name),

		GoPGVer: goPGVer,

		Entities: entities,
	}
}

// EntityData stores entity info for template
type EntityData struct {
	Name       string
	NamePlural string

	Imports []string

	VarName       string
	VarNamePlural string

	HasStatus bool
	HasPKs    bool
	PKs       []PKPair

	SortField string
	SortDir   string

	HasRelations bool
	Relations    []string

	Columns []AttributeData

	HasNotAddable bool
	NotAddable    []string

	HasNotUpdatable bool
	NotUpdatable    []string
}

// PackEntity packs mfd entity to template data
func PackEntity(entity mfd.Entity, options Options) EntityData {
	// base template entity - repo depends on int
	te := model.PackEntity(entity, model.Options{})

	hasStatus := false
	hasNotAddable := false
	hasNotUpdatable := false
	var notAddable []string
	var notUpdatable []string
	var pks []PKPair

	imports := mfd.NewSet()
	columns := make([]AttributeData, 0, len(te.Columns))

	for _, column := range te.Columns {
		// if has status - generate soft delete
		if mfd.IsStatus(column.DBName) {
			hasStatus = true
		}

		// if key - generate arg(s) for GetByID function
		if column.PrimaryKey {
			arg := util.LowerFirst(column.Name)
			if column.Name == util.ID {
				arg = "id"
			}
			if imp := mfd.Import(&column.Attribute, options.GoPGVer, options.CustomTypes); imp != "" {
				imports.Append(imp)
			}

			pks = append(pks, PKPair{
				Field: column.Name,
				Arg:   arg,
				Type:  column.GoType,
				Zero:  template.HTML(mfd.MakeZeroValue(column.GoType)),
			})
		}

		if !column.IsAddable() {
			hasNotAddable = true
			notAddable = append(notAddable, column.Name)
		}

		if !column.IsUpdatable() {
			hasNotUpdatable = true
			notUpdatable = append(notUpdatable, column.Name)
		}

		columns = append(columns, AttributeData{
			Name:      column.Name,
			Addable:   column.IsAddable(),
			Updatable: column.IsUpdatable(),
		})
	}

	// store all relation names for join field
	relNames := make([]string, len(te.Relations))
	for i := range te.Relations {
		relNames[i] = te.Relations[i].Name
	}

	// getting default sorts
	sortField, sortDir := sort(entity)

	// getting plural name for function name (eg CategoriesList)
	goNamePlural := mfd.MakePlural(te.Name)

	// getting var name for variable name (eg categories, newsList)
	varName := mfd.VarName(te.Name)
	varNamePlural := mfd.VarName(goNamePlural)
	if varName == varNamePlural {
		varNamePlural = fmt.Sprintf("%sList", varNamePlural)
	}

	return EntityData{
		Name:       te.Name,
		NamePlural: goNamePlural,

		Imports: imports.Elements(),

		VarName:       varName,
		VarNamePlural: varNamePlural,

		HasStatus: hasStatus,
		PKs:       pks,
		HasPKs:    len(pks) > 0,

		SortField: sortField,
		SortDir:   sortDir,

		Relations:    relNames,
		HasRelations: len(relNames) > 0,

		Columns:       columns,
		HasNotAddable: hasNotAddable,
		NotAddable:    notAddable,

		HasNotUpdatable: hasNotUpdatable,
		NotUpdatable:    notUpdatable,
	}
}

// AttributeData stores attribute info for template
type AttributeData struct {
	Name      string
	Addable   bool
	Updatable bool
}

func sort(entity mfd.Entity) (string, string) {
	presets := []SortPair{
		{"createdAt", "SortDesc"},
		{"title", "SortAsc"},
	}

	for _, preset := range presets {
		for _, attribute := range entity.Attributes {
			if attribute.DBName == preset.Field {
				return util.ColumnName(attribute.Name), preset.Dir
			}
		}
	}

	for _, attribute := range entity.Attributes {
		if attribute.PrimaryKey {
			return util.ColumnName(attribute.Name), "SortDesc"
		}
	}

	return "", ""
}
