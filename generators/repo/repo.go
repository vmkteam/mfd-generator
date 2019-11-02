package repo

import (
	"fmt"
	"html/template"

	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

type PKPair struct {
	Field string
	Arg   string
	Type  string
	Zero  template.HTML
}

type SortPair struct {
	Field, Dir string
}

type TemplatePackage struct {
	Package string

	Name         string
	ShortVarName string

	Entities []TemplateEntity
}

func NewTemplatePackage(namespace string, namespaces mfd.Namespaces, options Options) TemplatePackage {
	ns := namespaces.Namespace(namespace)
	entities := make([]TemplateEntity, len(ns.Entities))
	for i, entity := range ns.Entities {
		entities[i] = NewTemplateEntity(*entity)
	}

	name := util.CamelCased(util.Sanitize(namespace))

	return TemplatePackage{
		Package: options.Package,

		Name:         name,
		ShortVarName: mfd.ShortVarName(name),

		Entities: entities,
	}
}

type TemplateEntity struct {
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

	Columns         []TemplateColumn
	HasNotAddable   bool
	HasNotUpdatable bool
}

func NewTemplateEntity(entity mfd.Entity) TemplateEntity {
	// base template entity - repo depends on int
	te := model.NewTemplateEntity(entity, model.Options{})

	hasStatus := false
	hasNotAddable := false
	hasNotUpdatable := false
	var pks []PKPair

	var relations []string
	var columns []TemplateColumn

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

		columns = append(columns, TemplateColumn{
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

	return TemplateEntity{
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

type TemplateColumn struct {
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
