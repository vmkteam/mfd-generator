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

// RoutesNamespaceData stores package info
type RoutesNamespaceData struct {
	Entities []RoutesEntityData
}

// PackRoutesNamespace creates a package for template
func PackRoutesNamespace(namespaces []*mfd.VTNamespace) (RoutesNamespaceData, error) {
	var entities []RoutesEntityData

	for _, namespace := range namespaces {
		base, err := vt.PackNamespace(namespace, vt.Options{})
		if err != nil {
			return RoutesNamespaceData{}, err
		}

		for _, baseEntity := range base.Entities {
			if baseEntity.Mode == mfd.ModeReadOnly {
				continue
			}

			entity, err := PackRoutesEntity(baseEntity)
			if err != nil {
				return RoutesNamespaceData{}, err
			}
			entities = append(entities, entity)
		}
	}

	return RoutesNamespaceData{
		Entities: entities,
	}, nil
}

// RoutesEntityData stores routes struct info
type RoutesEntityData struct {
	vt.EntityData

	TerminalPath string

	JSName string
	PKs    []PKPair

	ModelColumns   []RoutesAttributeData
	ModelRelations []RoutesRelationData

	SummaryColumns   []RoutesAttributeData
	SummaryRelations []RoutesRelationData

	SearchColumns []RoutesAttributeData

	Params []RouterParamsData

	ReadOnly bool
}

// PackRoutesEntity creates an entity for routes template
func PackRoutesEntity(base vt.EntityData) (RoutesEntityData, error) {
	entity := RoutesEntityData{
		EntityData: base,

		TerminalPath: base.VTEntity.TerminalPath,
		JSName:       mfd.VarName(base.Name),

		ReadOnly: base.Mode == mfd.ModeReadOnlyWithTemplates,
	}

	for _, baseColumn := range base.ModelColumns {
		entity.ModelColumns = append(entity.ModelColumns, PackRoutesAttribute(baseColumn, entity))

		if baseColumn.Attribute.PrimaryKey {
			entity.PKs = append(entity.PKs, PKPair{
				JSName: mfd.VarName(baseColumn.Attribute.Name),
				JSType: template.HTML(mfd.MakeJSType(baseColumn.Attribute.GoType, baseColumn.IsArray)),
			})
		}
	}
	for _, baseRelation := range base.ModelRelations {
		entity.ModelRelations = append(entity.ModelRelations, PackRoutesRelation(baseRelation))
	}

	for _, baseColumn := range base.SummaryColumns {
		entity.SummaryColumns = append(entity.SummaryColumns, PackRoutesAttribute(baseColumn, entity))
	}
	for _, baseRelation := range base.SummaryRelations {
		entity.SummaryRelations = append(entity.SummaryRelations, PackRoutesRelation(baseRelation))
	}

	for _, baseColumn := range base.SearchColumns {
		entity.SearchColumns = append(entity.SearchColumns, PackRoutesAttribute(baseColumn, entity))
	}

	for _, baseParams := range base.Params {
		entity.Params = append(entity.Params, PackRoutesParams(baseParams))
	}

	return entity, nil
}

// RoutesAttributeData stores column info with attributes
type RoutesAttributeData struct {
	vt.AttributeData

	JSName string
	JSType template.HTML
	JSZero template.HTML
}

func PackRoutesAttribute(base vt.AttributeData, entity RoutesEntityData) RoutesAttributeData {
	jsType := mfd.MakeJSType(base.Attribute.GoType, base.IsArray)
	if base.IsParams {
		jsType = fmt.Sprintf("I%s%s", entity.Name, base.Name)
	}

	return RoutesAttributeData{
		AttributeData: base,

		JSName: mfd.VarName(base.VTAttribute.Name),
		JSType: template.HTML(jsType),
		JSZero: template.HTML(mfd.MakeJSZero(base.Attribute.GoType, base.IsArray)),
	}
}

// RoutesRelationData stores relation info with attributes
type RoutesRelationData struct {
	vt.RelationData

	JSName string
}

func PackRoutesRelation(base vt.RelationData) RoutesRelationData {
	return RoutesRelationData{
		RelationData: base,

		JSName: mfd.VarName(base.Name),
	}
}

type RouterParamsData struct {
	vt.ParamsData

	JSName string
}

func PackRoutesParams(base vt.ParamsData) RouterParamsData {
	return RouterParamsData{
		ParamsData: base,

		JSName: mfd.VarName(base.Name),
	}
}
