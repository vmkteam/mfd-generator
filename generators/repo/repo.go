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

	Entities []EntityData
}

// PackNamespace packs mfd namespace to template data
func PackNamespace(namespace *mfd.Namespace, options Options) NamespaceData {
	entities := make([]EntityData, len(namespace.Entities))
	for i, entity := range namespace.Entities {
		entities[i] = PackEntity(*entity)
	}

	name := util.CamelCased(util.Sanitize(namespace.Name))

	return NamespaceData{
		Package: options.Package,

		Name:         name,
		ShortVarName: mfd.ShortVarName(name),

		Entities: entities,
	}
}

// EntityData stores entity info for template
type EntityData struct {
	Name       string
	NamePlural string

	VarName       string
	VarNamePlural string

	HasStatus bool
	HasPKs    bool
	PKs       []PKPair

	SortField string
	SortDir   string

	HasRelations bool
	Relations    []string

	Columns         []AttributeData
	HasNotAddable   bool
	HasNotUpdatable bool
}

// PackEntity packs mfd entity to template data
func PackEntity(entity mfd.Entity) EntityData {
	// base template entity - repo depends on int
	te := model.PackEntity(entity, model.Options{})

	hasStatus := false
	hasNotAddable := false
	hasNotUpdatable := false
	var pks []PKPair

	var relations []string
	var columns []AttributeData

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
			pks = append(pks, PKPair{
				Field: column.Name,
				Arg:   arg,
				Type:  column.GoType,
				Zero:  template.HTML(mfd.MakeZeroValue(column.GoType)),
			})
		}

		if !column.IsAddable() {
			hasNotAddable = true
		}

		if !column.IsUpdatable() {
			hasNotUpdatable = true
		}

		columns = append(columns, AttributeData{
			Name:      column.Name,
			Addable:   column.IsAddable(),
			Updatable: column.IsUpdatable(),
		})
	}

	// store all relations for join field
	for _, relation := range te.Relations {
		relations = append(relations, relation.Name)
	}

	// getting default sorts
	sortField, sortDir := sort(entity)

	// getting plural name for function name (eg CategoriesList)
	schema, table := util.Split(entity.Table)
	goNamePlural := util.CamelCased(util.Sanitize(table))
	if schema != util.PublicSchema {
		goNamePlural = util.CamelCased(schema) + goNamePlural
	}

	// getting var name for variable name (eg categories, newsList)
	varName := mfd.VarName(te.Name)
	varNamePlural := mfd.VarName(goNamePlural)
	if varName == varNamePlural {
		varNamePlural = fmt.Sprintf("%sList", varNamePlural)
	}

	return EntityData{
		Name:       te.Name,
		NamePlural: goNamePlural,

		VarName:       varName,
		VarNamePlural: varNamePlural,

		HasStatus: hasStatus,
		PKs:       pks,
		HasPKs:    len(pks) > 0,

		SortField: sortField,
		SortDir:   sortDir,

		Relations:    relations,
		HasRelations: len(relations) > 0,

		Columns:         columns,
		HasNotAddable:   hasNotAddable,
		HasNotUpdatable: hasNotUpdatable,
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
