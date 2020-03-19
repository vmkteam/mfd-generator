package vt

import (
	"fmt"
	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
	base "github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/mfd"
	"html/template"
)

// this code is used to pack mdf to template

type PKPair struct {
	JSName string
	JSType template.HTML
}

// TemplatePackage stores package info
type TemplatePackage struct {
	Package string

	ModelPackage string

	HasImports bool
	Imports    []string

	Entities []TemplateEntity
}

// NewTemplatePackage creates a package for template
func NewTemplatePackage(namespace string, namespaces mfd.Namespaces, options Options) (TemplatePackage, error) {
	imports := mfd.NewSet()

	var models []TemplateEntity
	ns := namespaces.Namespace(namespace)
	for _, entity := range ns.Entities {
		// creating entity for template
		mdl, err := NewTemplateEntity(*entity)
		if err != nil {
			return TemplatePackage{}, err
		}

		models = append(models, mdl)
		// adding imports to uniq set
		for _, imp := range mdl.Imports {
			imports.Add(imp)
		}
	}

	return TemplatePackage{
		Package: options.Package,

		ModelPackage: options.ModelPackage,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}, nil
}

// TemplateEntity stores struct info
type TemplateEntity struct {
	mfd.VTEntity

	VarName      string
	ShortVarName string

	Imports []string

	PKs []PKPair

	ModelColumns      []TemplateColumn
	ModelRelations    []TemplateRelation
	HasModelRelations bool

	SummaryColumns      []TemplateColumn
	SummaryRelations    []TemplateRelation
	HasSummaryRelations bool

	SearchColumns []TemplateColumn

	Params    []TemplateParams
	HasParams bool
}

// NewTemplateEntity creates an entity for template
func NewTemplateEntity(entity mfd.Entity) (TemplateEntity, error) {
	imports := mfd.NewSet()

	tmpl := TemplateEntity{
		VTEntity: *entity.VTEntity,

		VarName:      mfd.VarName(entity.VTEntity.Name),
		ShortVarName: mfd.ShortVarName(entity.VTEntity.Name),

		PKs: []PKPair{},
	}

	// adding columns
	for _, vt := range entity.VTEntity.Attributes {
		// simple columns
		if vt.AttrName != "" {
			attr := entity.AttributeByName(vt.AttrName)
			// model column
			tmpl.ModelColumns = append(tmpl.ModelColumns, NewModelColumn(entity, *vt, *attr))
			if attr.ForeignKey != "" && !attr.IsArray {
				tmpl.ModelRelations = append(tmpl.ModelRelations, NewTemplateRelation(vt, attr))
			}

			// summary column
			if vt.Summary {
				tmpl.SummaryColumns = append(tmpl.SummaryColumns, NewSummaryColumn(entity, *vt, *attr))

				if attr.ForeignKey != "" && !attr.IsArray {
					tmpl.SummaryRelations = append(tmpl.SummaryRelations, NewTemplateRelation(vt, attr))
				}
			}

			// params column
			if attr.IsJSON() {
				tmpl.Params = append(tmpl.Params, NewTemplateParams(vt, attr))
			}

			// adding imports
			if imp := Import(attr); imp != "" {
				imports.Add(imp)
			}

			if attr.PrimaryKey {
				tmpl.PKs = append(tmpl.PKs, PKPair{
					JSName: mfd.VarName(attr.Name),
					JSType: template.HTML(mfd.MakeJSType(attr.GoType, attr.IsArray)),
				})
			}

			if mfd.IsStatus(attr.Name) {
				tmpl.ModelRelations = append(tmpl.ModelRelations, NewStatusRelation())
				if vt.Summary {
					tmpl.SummaryRelations = append(tmpl.SummaryRelations, NewStatusRelation())
				}
			}
		}
		// search columns
		if vt.Search {
			tmpl.SearchColumns = append(tmpl.SearchColumns, NewSearchColumn(entity, *vt))
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

// TemplateColumn stores column info
type TemplateColumn struct {
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

func NewModelColumn(entity mfd.Entity, vtAttr mfd.VTAttribute, attr mfd.Attribute) TemplateColumn {
	// model column as base
	baseColumn := base.NewTemplateColumn(entity, attr, base.Options{})

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

	column := TemplateColumn{
		VTAttribute: vtAttr,
		Attribute:   attr,

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
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(entity.VTEntity.Name), attr.Nullable())
		column.FromDBName, column.FromDBFunc = customFromIPConverter(column.Name, attr.Nullable())
		if attr.Nullable() {
			column.GoType = "*" + model.TypeString
		} else {
			column.GoType = model.TypeString
		}
	}

	return column
}

func NewSummaryColumn(entity mfd.Entity, vtAttr mfd.VTAttribute, attr mfd.Attribute) TemplateColumn {
	// model column as base
	baseColumn := base.NewTemplateColumn(entity, attr, base.Options{})

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(vtAttr.Name))

	column := TemplateColumn{
		VTAttribute: vtAttr,
		Attribute:   attr,

		Name:      util.ColumnName(vtAttr.Name),
		FieldName: baseColumn.Name,

		GoType:  baseColumn.GoType,
		IsArray: baseColumn.IsArray,

		Tag: template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}

	if attr.DBType == model.TypePGInet {
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(entity.VTEntity.Name), attr.Nullable())
		column.FromDBName, column.FromDBFunc = customFromIPConverter(column.Name, attr.Nullable())
		if attr.Nullable() {
			column.GoType = "*" + model.TypeString
		} else {
			column.GoType = model.TypeString
		}
	}

	return column
}

func NewSearchColumn(entity mfd.Entity, vtAttr mfd.VTAttribute) TemplateColumn {
	// search column as base
	var baseColumn base.SearchTemplateColumn
	var baseAttr mfd.Attribute
	if search := entity.SearchByName(vtAttr.SearchName); search != nil {
		baseColumn = base.NewCustomTemplateColumn(entity, *search, base.Options{})
		baseAttr = *search.Attribute
	} else if attr := entity.AttributeByName(vtAttr.SearchName); attr != nil {
		baseColumn = base.NewSearchTemplateColumn(entity, *attr, base.Options{})
		baseAttr = *attr
	}

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(vtAttr.Name))

	column := TemplateColumn{
		VTAttribute: vtAttr,
		Attribute:   baseAttr,

		Name:      vtAttr.Name,
		FieldName: baseColumn.Name,

		GoType:  baseColumn.GoType,
		IsArray: baseColumn.IsArray,

		Tag: template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}

	if baseAttr.DBType == model.TypePGInet {
		column.ToDBName, column.ToDBFunc = customToIPConverter(column.Name, mfd.ShortVarName(entity.VTEntity.Name)+"s", true)
		column.GoType = "*" + model.TypeString
	}

	return column
}

// TemplateRelation stores relation info
type TemplateRelation struct {
	Name      string
	FieldName string
	Type      string

	Tag template.HTML
}

func NewTemplateRelation(vtAttr *mfd.VTAttribute, attr *mfd.Attribute) TemplateRelation {
	name := util.ReplaceSuffix(util.ColumnName(attr.DBName), util.ID, "")

	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName(name))

	return TemplateRelation{
		Name:      name,
		FieldName: name,
		Type:      attr.ForeignKey,
		Tag:       template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}
}

func NewStatusRelation() TemplateRelation {
	tags := util.NewAnnotation()
	tags.AddTag("json", mfd.JSONName("status"))

	return TemplateRelation{
		Name:      "Status",
		FieldName: "Status",
		Type:      "Status",
		Tag:       template.HTML(fmt.Sprintf("`%s`", tags.String())),
	}
}

type TemplateParams struct {
	Name         string
	ShortVarName string
	FieldName    string
}

func NewTemplateParams(vtAttr *mfd.VTAttribute, attr *mfd.Attribute) TemplateParams {
	name := attr.GoType
	if name[0] == '*' {
		name = name[1:]
	}

	return TemplateParams{
		Name:         name,
		ShortVarName: mfd.ShortVarName(name),
		FieldName:    name,
	}
}

func Import(attribute *mfd.Attribute) string {
	return model.GoImport(attribute.DBType, attribute.Nullable(), false)
}

func customToIPConverter(name, entityShortName string, nullable bool) (template.HTML, template.HTML) {
	cVar := mfd.VarName(name) + "Field"
	tmpl := ""

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
	tmpl := ""

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
