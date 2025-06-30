package xml

import (
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

// this code used to convert entities from database to namespace in mfd project file

// PackEntity packs entity from db to mfd.Entity
func PackEntity(namespace string, entity model.Entity, existing *mfd.Entity, newCustomTypes mfd.CustomTypes) *mfd.Entity {
	var attribute *mfd.Attribute

	// processing all columns
	var attributes mfd.Attributes
	var searches mfd.Searches
	name := entity.GoName

	if existing != nil {
		attributes = existing.Attributes
		searches = existing.Searches
		name = existing.Name
	}

	hasAlias := false
	for _, column := range entity.Columns {
		if column.PGName == "alias" {
			hasAlias = true
		}
	}

	for _, column := range entity.Columns {
		_, ok := newCustomTypes.GoImport(mfd.Element(column.GoType), column.PGType)

		attributes, attribute = attributes.Merge(newAttribute(entity, column), ok)

		// do not add searches for existing columns
		if existing != nil && existing.AttributeByDBName(attribute.DBName, attribute.DBType) != nil {
			continue
		}

		// adding search if needed
		if column.IsPK {
			searches = searches.Append(newSearch(*attribute, mfd.SearchArray))

			if hasAlias {
				searches = searches.Append(newSearch(*attribute, mfd.SearchNotEquals))
			}
		}

		// making string searchable by like
		if !column.IsArray && column.GoType == model.TypeString && column.PGName != "alias" && column.PGName != "password" {
			searches = searches.Append(newSearch(*attribute, mfd.SearchILike))
		}
	}

	mfdEntity := &mfd.Entity{
		Name:       name,
		Namespace:  namespace,
		Table:      entity.PGFullName,
		Attributes: attributes,
		Searches:   searches,
	}

	return mfdEntity
}

func newAttribute(entity model.Entity, column model.Column) *mfd.Attribute {
	// special behaviour for statusId column
	if mfd.IsStatus(column.PGName) {
		return newStatusAttribute(column)
	}

	// processing foreign keys
	fkModel := ""
	if column.IsFK && column.Relation != nil {
		fkModel = column.Relation.GoType
	}

	// converting name to ID for PKs
	if column.IsPK && !entity.HasMultiplePKs() {
		column.GoName = util.ID
	}

	// making special type for json field: TableColumn, eg. UserParams, OrderCart...
	if column.PGType == model.TypePGJSON || column.PGType == model.TypePGJSONB {
		column.Type = entity.GoName + column.GoName
		if column.Nullable {
			column.Type = "*" + column.Type
		}
	}

	return &mfd.Attribute{
		Name:    column.GoName,
		DBName:  column.PGName,
		DBType:  column.PGType,
		GoType:  column.Type,
		IsArray: column.IsArray,

		PrimaryKey: column.IsPK,
		ForeignKey: fkModel,

		Addable:   addable(column),
		Updatable: updatable(column),
		Null:      nullable(column),
		Min:       0,
		Max:       column.MaxLen,
	}
}

func newSearch(attribute mfd.Attribute, searchType mfd.SearchType) *mfd.Search {
	return &mfd.Search{
		Name:       util.ColumnName(mfd.MakeSearchName(attribute.Name, searchType)),
		AttrName:   attribute.Name,
		SearchType: searchType,

		Attribute: &attribute,
	}
}

// nullable attribute logic here
func nullable(column model.Column) string {
	switch {
	case column.IsPK || column.Nullable:
		return mfd.NullableYes
	// case column.GoType == model.TypeString || column.IsFK:
	//	return mfd.NullableEmpty
	default:
		return mfd.NullableNo
	}
}

// addable attribute logic here
func addable(column model.Column) *bool {
	if column.PGName == "createdAt" || column.PGName == "modifiedAt" {
		return ptr(false)
	}

	return ptr(true)
}

// updatable attribute logic here
func updatable(column model.Column) *bool {
	if column.PGName == "createdAt" || column.PGName == "modifiedAt" {
		return ptr(false)
	}

	if column.IsPK {
		return ptr(false)
	}

	return ptr(true)
}

// default status column
func newStatusAttribute(column model.Column) *mfd.Attribute {
	addable := true
	updatable := true
	return &mfd.Attribute{
		Name:   column.GoName,
		DBName: column.PGName,

		DBType:  column.PGType,
		GoType:  column.Type,
		IsArray: false,

		PrimaryKey: false,
		Null:       mfd.NullableNo,
		Addable:    &addable,
		Updatable:  &updatable,
	}
}

func ptr[T comparable](t T) *T {
	return &t
}
