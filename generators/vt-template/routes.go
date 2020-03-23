package vttmpl

import (
	"fmt"
	"html/template"

	"github.com/vmkteam/mfd-generator/generators/vt"
	"github.com/vmkteam/mfd-generator/mfd"
)

// this code is used to pack mdf to template

type PKPair struct {
	JSName string
	JSType template.HTML
}

// NamespaceData stores package info
type TemplatePackage struct {
	Entities []TemplateEntity
}

// PackNamespace creates a package for template
func NewTemplatePackage(namespaces []*mfd.VTNamespace) (TemplatePackage, error) {
	var entities []TemplateEntity

	for _, namespace := range namespaces {
		base, err := vt.PackNamespace(namespace, vt.Options{})
		if err != nil {
			return TemplatePackage{}, err
		}

		for _, baseEntity := range base.Entities {
			if baseEntity.NoTemplates {
				continue
			}

			entity, err := NewTemplateEntity(baseEntity)
			if err != nil {
				return TemplatePackage{}, err
			}
			entities = append(entities, entity)
		}
	}

	return TemplatePackage{
		Entities: entities,
	}, nil
}

// EntityData stores struct info
type TemplateEntity struct {
	vt.EntityData

	TerminalPath string

	JSName string
	PKs    []PKPair

	ModelColumns   []TemplateColumn
	ModelRelations []TemplateRelation

	SummaryColumns   []TemplateColumn
	SummaryRelations []TemplateRelation

	SearchColumns []TemplateColumn

	Params []TemplateParams
}

// PackEntity creates an entity for template
func NewTemplateEntity(base vt.EntityData) (TemplateEntity, error) {
	entity := TemplateEntity{
		EntityData: base,

		TerminalPath: base.VTEntity.TerminalPath,
		JSName:       mfd.VarName(base.Name),
	}

	for _, baseColumn := range base.ModelColumns {
		entity.ModelColumns = append(entity.ModelColumns, NewTemplateColumn(baseColumn, entity))

		if baseColumn.Attribute.PrimaryKey {
			entity.PKs = append(entity.PKs, PKPair{
				JSName: mfd.VarName(baseColumn.Attribute.Name),
				JSType: template.HTML(mfd.MakeJSType(baseColumn.Attribute.GoType, baseColumn.IsArray)),
			})
		}
	}
	for _, baseRelation := range base.ModelRelations {
		entity.ModelRelations = append(entity.ModelRelations, NewTemplateRelation(baseRelation))
	}

	for _, baseColumn := range base.SummaryColumns {
		entity.SummaryColumns = append(entity.SummaryColumns, NewTemplateColumn(baseColumn, entity))
	}
	for _, baseRelation := range base.SummaryRelations {
		entity.SummaryRelations = append(entity.SummaryRelations, NewTemplateRelation(baseRelation))
	}

	for _, baseColumn := range base.SearchColumns {
		entity.SearchColumns = append(entity.SearchColumns, NewTemplateColumn(baseColumn, entity))
	}

	for _, baseParams := range base.Params {
		entity.Params = append(entity.Params, NewTemplateParams(baseParams))
	}

	return entity, nil
}

// AttributeData stores column info
type TemplateColumn struct {
	vt.AttributeData

	JSName string
	JSType template.HTML
	JSZero template.HTML
}

func NewTemplateColumn(base vt.AttributeData, entity TemplateEntity) TemplateColumn {
	jsType := mfd.MakeJSType(base.Attribute.GoType, base.IsArray)
	if base.IsParams {
		jsType = fmt.Sprintf("I%s%s", entity.Name, base.Name)
	}

	return TemplateColumn{
		AttributeData: base,

		JSName: mfd.VarName(base.VTAttribute.Name),
		JSType: template.HTML(jsType),
		JSZero: template.HTML(mfd.MakeJSZero(base.Attribute.GoType, base.IsArray)),
	}
}

// RelationData stores relation info
type TemplateRelation struct {
	vt.RelationData

	JSName string
}

func NewTemplateRelation(base vt.RelationData) TemplateRelation {
	return TemplateRelation{
		RelationData: base,

		JSName: mfd.VarName(base.Name),
	}
}

type TemplateParams struct {
	vt.ParamsData

	JSName string
}

func NewTemplateParams(base vt.ParamsData) TemplateParams {
	return TemplateParams{
		ParamsData: base,

		JSName: mfd.VarName(base.Name),
	}
}
