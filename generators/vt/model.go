package vt

import (
	"fmt"
	"html/template"

	base "github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

// this code is used to pack mdf to template

type PKPair struct {
	JSName string
	JSType template.HTML
}

// NamespaceData stores vt namespace info fro template
type NamespaceData struct {
	Package string

	ModelPackage string

	HasImports bool
	Imports    []string

	Entities []EntityData
}

// PackNamespace packs mfd vt namespace to template data
func PackNamespace(vtNamespace *mfd.VTNamespace, options Options) (NamespaceData, error) {
	imports := mfd.NewSet()

	models := make([]EntityData, 0, len(vtNamespace.Entities))
	for _, entity := range vtNamespace.Entities {
		if entity.Mode == mfd.ModeNone {
			continue
		}

		// creating entity for template
		mdl, err := PackEntity(*entity, options)
		if err != nil {
			return NamespaceData{}, err
		}

		models = append(models, mdl)
		// adding imports to uniq set
		for _, imp := range mdl.Imports {
			imports.Add(imp)
		}
	}

	return NamespaceData{
		Package: options.Package,

		ModelPackage: options.ModelPackage,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}, nil
}

// EntityData stores mfd vt entity info
type EntityData struct {
	mfd.VTEntity

	VarName      string
	ShortVarName string

	Imports []string

	PKs []PKPair

	ModelColumns      []AttributeData
	ModelRelations    []RelationData
	HasModelRelations bool

	SummaryColumns      []AttributeData
	SummaryRelations    []RelationData
	HasSummaryRelations bool

	SearchColumns []AttributeData

	Params    []ParamsData
	HasParams bool
}

// PackEntity packs mfd vt entity to template data
func PackEntity(vtEntity mfd.VTEntity, options Options) (EntityData, error) {
	imports := mfd.NewSet()

	tmpl := EntityData{
		VTEntity: vtEntity,

		VarName:      mfd.VarName(vtEntity.Name),
		ShortVarName: mfd.ShortVarName(vtEntity.Name),

		PKs: []PKPair{},
	}

	// adding columns
	for _, vtAttr := range vtEntity.Attributes {
		// simple columns
		if vtAttr.AttrName != "" {
			// corresponding entity attr
			attr := vtAttr.Attribute

			// model column
			tmpl.ModelColumns = append(tmpl.ModelColumns, PackAttribute(vtEntity, *vtAttr))
			if attr.ForeignKey != "" && !attr.IsArray {
				tmpl.ModelRelations = append(tmpl.ModelRelations, PackRelation(vtAttr))
			}

			// summary column
			if vtAttr.Summary {
				tmpl.SummaryColumns = append(tmpl.SummaryColumns, PackSummaryAttribute(vtEntity, *vtAttr))

				if attr.ForeignKey != "" && !attr.IsArray {
					tmpl.SummaryRelations = append(tmpl.SummaryRelations, PackRelation(vtAttr))
				}
			}

			// params column
			if attr.IsJSON() {
				tmpl.Params = append(tmpl.Params, PackParams(vtAttr))
			}

			// adding imports
			if imp := mfd.Import(attr, options.GoPGVer, options.CustomTypes); imp != "" {
				imports.Add(imp)
			}

			if attr.PrimaryKey {
				tmpl.PKs = append(tmpl.PKs, PKPair{
					JSName: mfd.VarName(attr.Name),
					JSType: template.HTML(mfd.MakeJSType(attr.GoType, attr.IsArray)),
				})
			}

			if mfd.IsStatus(attr.Name) {
				tmpl.ModelRelations = append(tmpl.ModelRelations, PackStatusRelation())
				if vtAttr.Summary {
					tmpl.SummaryRelations = append(tmpl.SummaryRelations, PackStatusRelation())
				}
			}
		}
		// search columns
		if vtAttr.Search {
			sa, err := PackSearchAttribute(vtEntity, *vtAttr)
			if err != nil {
				return EntityData{}, fmt.Errorf("pack search attribute, err=%w", err)
			}
			tmpl.SearchColumns = append(tmpl.SearchColumns, sa)
		}
	}

	tmpl.HasModelRelations = len(tmpl.ModelRelations) > 0
	tmpl.HasSummaryRelations = len(tmpl.SummaryRelations) > 0
	tmpl.HasParams = len(tmpl.Params) > 0

	if imports.Len() > 0 {
		tmpl.Imports = imports.Elements()
	}

	return tmpl, nil
}

// AttributeData stores vt attribute info
type AttributeData struct {
	VTAttribute mfd.VTAttribute
	Attribute   mfd.Attribute

	Name      string
	FieldName string
	GoType    string
	IsArray   bool

	Tag     template.HTML
	Comment template.HTML

	NilCheck bool

	IsParams   bool
	ParamsName string

	ToDBFunc   template.HTML
	ToDBName   template.HTML
	FromDBFunc template.HTML
	FromDBName template.HTML
}

// PackAttribute packs mfd vt attribute to template data
func PackAttribute(vtEntity mfd.VTEntity, vtAttr mfd.VTAttribute) AttributeData {
	// corresponding entity and attribute
	entity := vtEntity.Entity
	attr := vtAttr.Attribute

	// model column as base
	baseColumn := base.PackAttribute(*entity, *attr, base.Options{})

	// adding tags
	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(vtAttr.Name))

	if vtAttr.Required {
		tags.AddTag("validate", "required")
	}
	if attr.Nullable() && (vtAttr.Validate != "" || vtAttr.MaxValue != 0 || vtAttr.MinValue != 0) {
		tags.AddTag("validate", "omitempty")
	}

	if vtAttr.Validate != "" {
		tags.AddTag("validate", vtAttr.Validate)
	}
	if vtAttr.MaxValue != 0 {
		tags.AddTag("validate", fmt.Sprintf("max=%d", vtAttr.MaxValue))
	}
	if vtAttr.MinValue != 0 {
		tags.AddTag("validate", fmt.Sprintf("min=%d", vtAttr.MinValue))
	}

	column := AttributeData{
		VTAttribute: vtAttr,
		Attribute:   *attr,

		Name:      util.ColumnName(vtAttr.Name),
		FieldName: baseColumn.Name,

		GoType:  baseColumn.GoType,
		IsArray: baseColumn.IsArray,

		Tag: template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}

	// Adding ParamsLogic to column
	if attr.IsJSON() {
		name := attr.GoType
		if name[0] == '*' {
			name = name[1:]
		}
		column.IsParams = true
		column.ParamsName = name
		column.NilCheck = attr.Nullable()
	}

	if attr.DBType == model.TypePGInet {
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(vtEntity.Name), attr.Nullable())
		column.FromDBName, column.FromDBFunc = customFromIPConverter(column.Name, attr.Nullable())
		if attr.Nullable() {
			column.GoType = "*" + model.TypeString
		} else {
			column.GoType = model.TypeString
		}
	}

	return column
}

// PackSummaryAttribute packs mfd vt attribute to summary template data
func PackSummaryAttribute(vtEntity mfd.VTEntity, vtAttr mfd.VTAttribute) AttributeData {
	// corresponding entity and attribute
	entity := vtEntity.Entity
	attr := vtAttr.Attribute

	// model column as base
	baseColumn := base.PackAttribute(*entity, *attr, base.Options{})

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(vtAttr.Name))

	column := AttributeData{
		VTAttribute: vtAttr,
		Attribute:   *attr,

		Name:      util.ColumnName(vtAttr.Name),
		FieldName: baseColumn.Name,

		GoType:  baseColumn.GoType,
		IsArray: baseColumn.IsArray,

		Tag: template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}

	if attr.DBType == model.TypePGInet {
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(vtEntity.Name), attr.Nullable())
		column.FromDBName, column.FromDBFunc = customFromIPConverter(column.Name, attr.Nullable())
		if attr.Nullable() {
			column.GoType = "*" + model.TypeString
		} else {
			column.GoType = model.TypeString
		}
	}

	return column
}

// PackSearchAttribute packs mfd vt attribute to search template data
func PackSearchAttribute(vtEntity mfd.VTEntity, vtAttr mfd.VTAttribute) (AttributeData, error) {
	// corresponding entity
	entity := vtEntity.Entity

	// search column as base
	var (
		baseColumn base.SearchAttributeData
		baseAttr   mfd.Attribute
		err        error
	)

	if search := entity.SearchByName(vtAttr.SearchName); search != nil {
		baseColumn, err = base.CustomSearchAttribute(*entity, *search, base.Options{})
		if err != nil {
			return AttributeData{}, fmt.Errorf("custom search attribute, err=%w", err)
		}
		baseAttr = *search.Attribute
	} else if attr := entity.AttributeByName(vtAttr.SearchName); attr != nil {
		baseColumn = base.PackSearchAttribute(*entity, *attr, base.Options{})
		baseAttr = *attr
	}

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(vtAttr.Name))

	column := AttributeData{
		VTAttribute: vtAttr,
		Attribute:   baseAttr,

		Name:      vtAttr.Name,
		FieldName: baseColumn.Name,

		GoType:  baseColumn.GoType,
		IsArray: baseColumn.IsArray,

		Tag: template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}

	if baseAttr.DBType == model.TypePGInet {
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(vtEntity.Name)+"s", true)
		column.GoType = "*" + model.TypeString
	}

	return column, nil
}

// RelationData stores relation info
type RelationData struct {
	Name      string
	FieldName string
	Type      string

	Tag template.HTML
}

// PackRelation packs mfd vt attribute to relation template data
func PackRelation(vtAttr *mfd.VTAttribute) RelationData {
	name := util.ReplaceSuffix(util.ColumnName(vtAttr.Attribute.DBName), util.ID, "")

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(name))

	return RelationData{
		Name:      name,
		FieldName: name,
		Type:      vtAttr.Attribute.ForeignKey,
		Tag:       template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}
}

// PackStatusRelation creates relation for status vt attribute
func PackStatusRelation() RelationData {
	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName("status"))

	return RelationData{
		Name:      "Status",
		FieldName: "Status",
		Type:      "Status",
		Tag:       template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}
}

// ParamsData stores vt params info
type ParamsData struct {
	Name         string
	ShortVarName string
	FieldName    string
}

// PackParams packs mfd vt attribute to vt params template data
func PackParams(vtAttr *mfd.VTAttribute) ParamsData {
	name := vtAttr.Attribute.GoType
	if name[0] == '*' {
		name = name[1:]
	}

	return ParamsData{
		Name:         name,
		ShortVarName: mfd.ShortVarName(name),
		FieldName:    name,
	}
}

func customToIPConverter(name, entityShortName string, nullable bool) (template.HTML, template.HTML) {
	cVar := mfd.VarName(name) + "Field"
	var tmpl string

	if nullable {
		tmpl = fmt.Sprintf(`
			var %s *net.IP
			if %s.%s != nil {
				%sProxy := net.ParseIP(*%s.%s)
				%s = &%sProxy
			}
		`, cVar, entityShortName, name, cVar, entityShortName, name, cVar, cVar)
	} else {
		tmpl = fmt.Sprintf(`
			%s := net.ParseIP(%s.%s)
		`, cVar, entityShortName, name)
	}

	return template.HTML(cVar), template.HTML(tmpl)
}

func customFromIPConverter(name string, nullable bool) (template.HTML, template.HTML) {
	cVar := mfd.VarName(name) + "Field"
	var tmpl string

	if nullable {
		tmpl = fmt.Sprintf(`
			var %s *string 
			if in.%s != nil {
				%sProxy := in.%s.String()
				%s = &%sProxy
			}
		`, cVar, name, cVar, name, cVar, cVar)
	} else {
		tmpl = fmt.Sprintf(`
			%s := in.%s.String() 
		`, cVar, name)
	}

	return template.HTML(cVar), template.HTML(tmpl)
}
