package vt

import (
	"fmt"
	"html/template"

	"github.com/vmkteam/mfd-generator/generators/model"
	base "github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// ServiceNamespaceData stores namespace info
type ServiceNamespaceData struct {
	Package         string
	ModelPackage    string
	EmbedLogPackage string

	Name    string
	VarName string

	HasImports bool
	Imports    []string

	Entities []ServiceEntityData
}

// PackServiceNamespace packs mfd vt namespace to template data
func PackServiceNamespace(namespace *mfd.VTNamespace, options Options) (ServiceNamespaceData, error) {
	imports := mfd.NewSet()
	entities := make([]ServiceEntityData, 0, len(namespace.Entities))
	for _, entity := range namespace.Entities {
		if entity.Mode == mfd.ModeNone {
			continue
		}

		packed, err := PackServiceEntity(*entity, options)
		if err != nil {
			return ServiceNamespaceData{}, fmt.Errorf("packing service entity: %w", err)
		}
		entities = append(entities, packed)
		for _, imp := range packed.Imports {
			imports.Append(imp)
		}
	}

	name := util.CamelCased(util.Sanitize(namespace.Name))

	return ServiceNamespaceData{
		Package:         options.Package,
		ModelPackage:    options.ModelPackage,
		EmbedLogPackage: options.EmbedLogPackage,

		Name:    name,
		VarName: mfd.VarName(name),

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: entities,
	}, nil
}

func (tp ServiceNamespaceData) Raw(s string) template.HTML {
	return template.HTML(s)
}

// ServiceEntityData stores entity info
type ServiceEntityData struct {
	Name          string
	NamePlural    string
	VarName       string
	VarNamePlural string
	ShortVarName  string

	Imports []string

	HasSortColumns bool
	SortColumns    []string

	PKs []base.PKPair

	HasAlias   bool
	PKSearches []base.PKPair
	AliasField string
	AliasArg   string

	HasRelations    bool
	Relations       []ServiceRelationData
	UniqueRelations []ServiceRelationData

	ReadOnly bool
}

// PackServiceEntity packs mfd vt entity to template data
func PackServiceEntity(vtEntity mfd.VTEntity, options Options) (ServiceEntityData, error) {
	baseEntity := base.PackEntity(*vtEntity.Entity, base.Options{
		Package:     options.Package,
		GoPGVer:     options.GoPGVer,
		CustomTypes: options.CustomTypes,
	})

	var relations []ServiceRelationData
	var uniqueRelations []ServiceRelationData
	var sortColumns []string
	foreignKeys := make(map[string]struct{})
	for _, vtAttr := range vtEntity.Attributes {
		if vtAttr.AttrName != "" {
			if !vtAttr.Attribute.IsArray && vtAttr.Summary {
				sortColumns = append(sortColumns, vtAttr.Attribute.Name)
			}

			if vtAttr.Attribute.ForeignKey != "" && vtAttr.Attribute.ForeignEntity != nil {
				serviceRelationData := PackServiceRelationData(*vtAttr, *vtAttr.Attribute.ForeignEntity)
				relations = append(relations, serviceRelationData)
				if _, ok := foreignKeys[vtAttr.Attribute.ForeignEntity.Namespace]; !ok {
					foreignKeys[vtAttr.Attribute.ForeignEntity.Namespace] = struct{}{}
					uniqueRelations = append(uniqueRelations, serviceRelationData)
				}
			}
		}
	}

	// setting search for alias unique
	var pkSearches []base.PKPair
	var aliasField, aliasArg string
	for _, vtAttr := range vtEntity.Attributes {
		if vtAttr.AttrName == "" {
			continue
		}
		attr := vtAttr.Attribute

		if attr.DBName == "alias" {
			aliasField, aliasArg = attr.Name, attr.Name
		}

		if attr.PrimaryKey {
			search := vtEntity.Entity.SearchByAttrName(attr.Name, mfd.SearchNotEquals)
			if search == nil {
				continue
			}

			column, err := model.CustomSearchAttribute(*vtEntity.Entity, *search, model.Options{})
			if err != nil {
				return ServiceEntityData{}, fmt.Errorf("custom search attribute: %w", err)
			}
			pkSearches = append(pkSearches, base.PKPair{
				Field: attr.Name,
				Arg:   column.Name,
			})
		}
	}

	return ServiceEntityData{
		Name:          baseEntity.Name,
		NamePlural:    baseEntity.NamePlural,
		VarName:       baseEntity.VarName,
		VarNamePlural: baseEntity.VarNamePlural,
		ShortVarName:  mfd.ShortVarName(baseEntity.Name),

		HasSortColumns: len(sortColumns) > 0,
		SortColumns:    sortColumns,

		PKs:     baseEntity.PKs,
		Imports: baseEntity.Imports,

		HasAlias:   aliasField != "",
		PKSearches: pkSearches,
		AliasField: aliasField,
		AliasArg:   aliasArg,

		HasRelations:    len(relations) > 0,
		Relations:       relations,
		UniqueRelations: uniqueRelations,

		ReadOnly: vtEntity.Mode == mfd.ModeReadOnly || vtEntity.Mode == mfd.ModeReadOnlyWithTemplates,
	}, nil
}

// ServiceRelationData stores relation info
type ServiceRelationData struct {
	Name      string
	NameSpace string
	FK        string
	PluralFK  string
	JSONName  string
	Nullable  bool
	IsArray   bool
}

// PackServiceRelationData packs mfd vt attribute to relation template data
func PackServiceRelationData(vtAttr mfd.VTAttribute, foreign mfd.Entity) ServiceRelationData {
	attr := vtAttr.Attribute

	baseRelation := model.PackRelation(*attr, model.Options{})

	return ServiceRelationData{
		Name:      attr.Name,
		JSONName:  mfd.JSONName(vtAttr.Name),
		FK:        baseRelation.ForeignEntity.Name,
		PluralFK:  mfd.MakePlural(util.CamelCased(baseRelation.ForeignEntity.Name)),
		NameSpace: foreign.Namespace,
		Nullable:  attr.Nullable(),
		IsArray:   attr.IsArray,
	}
}
