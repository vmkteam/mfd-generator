package vt

import (
	"html/template"

	"github.com/vmkteam/mfd-generator/generators/model"
	base "github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/mfd"
)

type ServiceTemplatePackage struct {
	Package      string
	ModelPackage string

	Name    string
	VarName string

	Entities []ServiceTemplateEntity
}

func NewServiceTemplatePackage(namespace string, namespaces mfd.Namespaces, options Options) ServiceTemplatePackage {
	basePkg := base.NewTemplatePackage(namespace, namespaces, base.Options{})

	ns := namespaces.Namespace(namespace)
	entities := make([]ServiceTemplateEntity, len(ns.Entities))
	for i, entity := range ns.Entities {
		entities[i] = NewServiceTemplateEntity(*entity)
	}

	return ServiceTemplatePackage{
		Package:      options.Package,
		ModelPackage: options.ModelPackage,

		Name:    basePkg.Name,
		VarName: mfd.VarName(basePkg.Name),

		Entities: entities,
	}
}

func (tp ServiceTemplatePackage) Raw(s string) template.HTML {
	return template.HTML(s)
}

type ServiceTemplateEntity struct {
	Name          string
	NamePlural    string
	VarName       string
	VarNamePlural string
	ShortVarName  string

	HasSortColumns bool
	SortColumns    []string

	PKs []base.PKPair

	HasAlias   bool
	PKSearches []base.PKPair
	AliasField string
	AliasArg   string

	HasRelations bool
	Relations    []ServiceTemplateRelation
}

func NewServiceTemplateEntity(entity mfd.Entity) ServiceTemplateEntity {
	baseEntity := base.NewTemplateEntity(entity)

	var relations []ServiceTemplateRelation
	var sortColumns []string
	for _, vtAttr := range entity.VTEntity.Attributes {
		attr := entity.AttributeByName(vtAttr.AttrName)
		if attr != nil && !attr.IsArray && vtAttr.Summary {
			sortColumns = append(sortColumns, attr.Name)
		}

		if attr != nil && attr.ForeignKey != "" {
			relations = append(relations, NewServiceTemplateRelation(*vtAttr, *attr))
		}
	}

	// setting search for alias unique
	var pkSearches []base.PKPair
	var aliasField, aliasArg string
	for _, attr := range entity.Attributes {
		if attr.DBName == "alias" {
			aliasField, aliasArg = attr.Name, attr.Name
		}

		if attr.PrimaryKey {
			search := entity.SearchByAttrName(attr.Name, mfd.SearchNotEquals)
			if search == nil {
				continue
			}

			column := model.NewCustomTemplateColumn(entity, *search, model.Options{})
			pkSearches = append(pkSearches, base.PKPair{
				Field: attr.Name,
				Arg:   column.Name,
			})
		}
	}

	return ServiceTemplateEntity{
		Name:          baseEntity.Name,
		NamePlural:    baseEntity.NamePlural,
		VarName:       baseEntity.VarName,
		VarNamePlural: baseEntity.VarNamePlural,
		ShortVarName:  mfd.ShortVarName(baseEntity.Name),

		HasSortColumns: len(sortColumns) > 0,
		SortColumns:    sortColumns,

		PKs: baseEntity.PKs,

		HasAlias:   aliasField != "",
		PKSearches: pkSearches,
		AliasField: aliasField,
		AliasArg:   aliasArg,

		HasRelations: len(relations) > 0,
		Relations:    relations,
	}
}

type ServiceTemplateRelation struct {
	Name     string
	FK       string
	JSONName string
	Nullable bool
}

func NewServiceTemplateRelation(vtAttr mfd.VTAttribute, attr mfd.Attribute) ServiceTemplateRelation {
	baseRelation := model.NewTemplateRelation(attr, model.Options{})

	return ServiceTemplateRelation{
		Name:     attr.Name,
		JSONName: mfd.JSONName(vtAttr.Name),
		FK:       baseRelation.ForeignEntity.Name,
		Nullable: attr.Nullable(),
	}
}
