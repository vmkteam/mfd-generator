package mfd

import (
	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

// this code used to make array of model.Entity from mfd file

// NewGennaEntities convert Namespace to genna entities
func NewGennaEntities(pack *Namespace) []model.Entity {
	result := make([]model.Entity, len(pack.Entities))
	index := map[string]int{}
	for i, e := range pack.Entities {
		entity := NewGennaEntity(e)

		result[i] = entity
		index[util.Join(entity.PGSchema, entity.PGName)] = i
	}

	// making links
	for i, e := range result {
		if len(e.Relations) == 0 {
			continue
		}
		// linking relation to entity
		for j, r := range e.Relations {
			if target, ok := index[util.Join(r.TargetPGSchema, r.TargetPGName)]; ok {
				result[i].Relations[j].AddEntity(&result[target])
			}
		}
	}

	return result
}

func NewGennaEntity(entity *Entity) model.Entity {
	// creating genna entity
	schema, table := util.Split(entity.Table)
	e := model.NewEntity(util.Sanitize(schema), util.Sanitize(table), nil, nil)

	// adding columns & relations from mfd attributes
	for _, a := range entity.Attributes {
		column := NewGennaColumn(e, a)

		if a.ForeignKey != "" {
			relation := NewGennaRelation(a)
			column.AddRelation(&relation)

			e.AddRelation(relation)
		}

		e.AddColumn(column)
	}

	return e
}

func NewGennaRelation(attr *Attribute) model.Relation {
	schema, table := util.Split(attr.ForeignKey)
	return model.NewRelation([]string{attr.DBName}, schema, table)
}

func NewGennaColumn(entity model.Entity, attr *Attribute) model.Column {
	// getting simple go type (for search mostly)
	goType, err := model.GoType(attr.DBType)
	if err != nil {
		goType = model.TypeInterface
	}

	// adding dimensions for arrays
	dims := 0
	if attr.IsArray {
		dims = 1
	}

	nullable := attr.Nullable() && !attr.PrimaryKey

	return model.Column{
		GoName:     util.ColumnName(attr.Name),
		PGName:     attr.DBName,
		Type:       attr.GoType,
		GoType:     goType,
		PGType:     attr.DBType,
		Nullable:   nullable,
		IsArray:    attr.IsArray,
		Dimensions: dims,
		IsPK:       attr.PrimaryKey,
		IsFK:       attr.ForeignKey != "",
		Import:     model.GoImport(attr.DBType, nullable, false, 8),
		MaxLen:     attr.Max,
	}
}
