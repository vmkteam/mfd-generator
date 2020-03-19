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

// TemplatePackage stores package info
type TemplatePackage struct {
	Entities []TemplateEntity
}

// NewTemplatePackage creates a package for template
func NewTemplatePackage(namespaces mfd.Namespaces, options Options) (TemplatePackage, error) {
	var entities []TemplateEntity

	for _, namespace := range namespaces {
		base, err := vt.NewTemplatePackage(namespace.Name, namespaces, vt.Options{})
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

// TemplateEntity stores struct info
type TemplateEntity struct {
	vt.TemplateEntity

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

// NewTemplateEntity creates an entity for template
func NewTemplateEntity(base vt.TemplateEntity) (TemplateEntity, error) {
	entity := TemplateEntity{
		TemplateEntity: base,

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

// TemplateColumn stores column info
type TemplateColumn struct {
	vt.TemplateColumn

	JSName string
	JSType template.HTML
	JSZero template.HTML
}

func NewTemplateColumn(base vt.TemplateColumn, entity TemplateEntity) TemplateColumn {
	jsType := mfd.MakeJSType(base.Attribute.GoType, base.IsArray)
	if base.IsParams {
		jsType = fmt.Sprintf("I%s%s", entity.Name, base.Name)
	}

	return TemplateColumn{
		TemplateColumn: base,

		JSName: mfd.VarName(base.VTAttribute.Name),
		JSType: template.HTML(jsType),
		JSZero: template.HTML(mfd.MakeJSZero(base.Attribute.GoType, base.IsArray)),
	}
}

// TemplateRelation stores relation info
type TemplateRelation struct {
	vt.TemplateRelation

	JSName string
}

func NewTemplateRelation(base vt.TemplateRelation) TemplateRelation {
	return TemplateRelation{
		TemplateRelation: base,

		JSName: mfd.VarName(base.Name),
	}
}

type TemplateParams struct {
	vt.TemplateParams

	JSName string
}

func NewTemplateParams(base vt.TemplateParams) TemplateParams {
	return TemplateParams{
		TemplateParams: base,

		JSName: mfd.VarName(base.Name),
	}
}
