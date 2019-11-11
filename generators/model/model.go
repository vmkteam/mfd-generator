package model

import (
	"fmt"
	"html/template"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

// this code is used to pack mdf to template

// TemplatePackage stores package info
type TemplatePackage struct {
	Package string

	HasImports bool
	Imports    []string

	Entities []TemplateEntity
}

// NewTemplatePackage creates a package for template
func NewTemplatePackage(namespaces mfd.Namespaces, options Options) TemplatePackage {
	imports := mfd.NewSet()

	var models []TemplateEntity
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			// creating entity for template
			mdl := NewTemplateEntity(*entity, options)
			models = append(models, mdl)
			// adding imports to uniq set
			for _, imp := range mdl.Imports {
				imports.Add(imp)
			}

		}
	}

	return TemplatePackage{
		Package: options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}
}

// TemplateEntity stores struct info
type TemplateEntity struct {
	mfd.Entity

	ShortVarName string

	Tag template.HTML

	NoAlias bool
	Alias   string

	Imports []string

	Columns []TemplateColumn

	HasRelations bool
	Relations    []TemplateRelation
}

// NewTemplateEntity creates an entity for template
func NewTemplateEntity(entity mfd.Entity, options Options) TemplateEntity {
	imports := mfd.NewSet()
	var columns []TemplateColumn
	var relations []TemplateRelation

	// adding columns
	for _, attribute := range entity.Attributes {
		column := NewTemplateColumn(entity, *attribute, options)
		columns = append(columns, column)

		// adding imports to uniq set
		if column.Import != "" {
			imports.Add(column.Import)
		}

		// adding relation from column
		if attribute.ForeignKey != "" && !attribute.IsArray {
			relations = append(relations, NewTemplateRelation(*attribute, options))
		}
	}

	// adding annotations for go-pg to column
	tagName := tagName(options)
	tags := util.NewAnnotation()
	tags.AddTag(tagName, util.Quoted(entity.Table, true))
	tags.AddTag(tagName, fmt.Sprintf("alias:%s", util.DefaultAlias))
	tags.AddTag("pg", ",discard_unknown_columns")

	return TemplateEntity{
		Entity: entity,

		ShortVarName: mfd.ShortVarName(entity.Name),

		// avoid escaping
		Tag:   template.HTML(fmt.Sprintf("`%s`", tags.String())),
		Alias: util.DefaultAlias,

		Imports: imports.Elements(),

		Columns: columns,

		HasRelations: len(relations) > 0,
		Relations:    relations,
	}
}

// TemplateColumn stores column info
type TemplateColumn struct {
	mfd.Attribute

	Name   string
	Type   string
	Import string

	Tag     template.HTML
	Comment template.HTML
}

// NewTemplateColumn creates a column for template
func NewTemplateColumn(entity mfd.Entity, attribute mfd.Attribute, options Options) TemplateColumn {
	comment := ""
	tagName := tagName(options)
	tags := util.NewAnnotation()
	tags.AddTag(tagName, attribute.DBName)

	// pk tag
	if attribute.PrimaryKey {
		tags.AddTag(tagName, "pk")
	}

	// types tag
	if attribute.DBType == model.TypePGHstore {
		tags.AddTag(tagName, "hstore")
	} else if attribute.IsArray {
		tags.AddTag(tagName, "array")
	}
	if attribute.DBType == model.TypePGUuid {
		tags.AddTag(tagName, "type:uuid")
	}

	// nullable tag
	if !attribute.Nullable() && !attribute.PrimaryKey {
		if options.GoPgVer == 9 {
			tags.AddTag(tagName, "use_zero")
		} else {
			tags.AddTag(tagName, "notnull")
		}
	}

	// go type
	goType, err := model.GoType(attribute.DBType)
	if err != nil {
		goType = model.TypeInterface
	}

	// mark unknown types as interface & unsupported
	if goType == model.TypeInterface {
		comment = "// unsupported"
		tags = util.NewAnnotation().AddTag(tagName, "-")
	}

	// fix pointer in case of inconsistency
	attribute.GoType = fixPointer(attribute)

	return TemplateColumn{
		Attribute: attribute,

		Type:   goType,
		Name:   util.ColumnName(attribute.Name),
		Import: model.GoImport(attribute.DBType, attribute.Nullable(), false),

		Tag:     template.HTML(fmt.Sprintf("`%s`", tags.String())),
		Comment: template.HTML(comment),
	}
}

// TemplateRelation stores relation info
type TemplateRelation struct {
	mfd.Attribute

	Name string
	Type string

	Tag     template.HTML
	Comment template.HTML
}

// NewTemplateRelation creates relation for template
func NewTemplateRelation(relation mfd.Attribute, options Options) TemplateRelation {
	// adding go-pg's fk annotation
	tags := util.NewAnnotation().AddTag("pg", "fk:"+relation.DBName)
	comment := ""

	// getting pk in foreign table
	var fkFields []string
	if relation.ForeignEntity != nil {
		pks := relation.ForeignEntity.PKs()
		for _, pk := range pks {
			fkFields = append(fkFields, pk.DBName)
		}
	}

	if len(fkFields) > 1 {
		tagName := tagName(options)
		tags.AddTag(tagName, "-")
		comment = "// unsupported"
	}

	return TemplateRelation{
		Attribute: relation,

		// ObjectID -> Object, UserID -> User
		Name: util.ReplaceSuffix(util.ColumnName(relation.DBName), util.ID, ""),
		Type: relation.ForeignKey,

		Tag:     template.HTML(fmt.Sprintf("`%s`", tags.String())),
		Comment: template.HTML(comment),
	}
}

func tagName(options Options) string {
	if options.GoPgVer == 9 {
		return "pg"
	}
	return "sql"
}

func fixPointer(attribute mfd.Attribute) string {
	// basic
	if !attribute.Nullable() || attribute.PrimaryKey {
		return attribute.GoType
	}

	// type opts
	if attribute.IsMap() || attribute.IsArray {
		return attribute.GoType
	}

	if attribute.GoType[0] != '*' {
		return "*" + attribute.GoType
	}

	return attribute.GoType
}
