package xml

import (
	"strings"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
	"github.com/vmkteam/mfd-generator/mfd"
)

// this code used to convert entities from database to namespace in mfd project file

// PackEntity packs entity from db to mfd.Entity
func PackEntity(namespace string, entity model.Entity) *mfd.Entity {
	// processing all columns
	attributes := mfd.Attributes{}
	searches := mfd.Searches{}

	hasAlias := false
	for _, column := range entity.Columns {
		if column.PGName == "alias" {
			hasAlias = true
		}
	}

	for _, column := range entity.Columns {
		attribute := newAttribute(entity, column)

		attributes = append(attributes, attribute)

		// adding search if needed
		if column.IsPK {
			searches = append(searches, newSearch(column, *attribute, mfd.SearchArray))
			if hasAlias {
				searches = append(searches, newSearch(column, *attribute, mfd.SearchNotEquals))
			}
		}

		// making string searchable by like
		if !column.IsArray && column.GoType == model.TypeString && column.PGName != "alias" && column.PGName != "password" {
			searches = append(searches, newSearch(column, *attribute, mfd.SearchILike))
		}
	}

	mfdEntity := &mfd.Entity{
		Name:       entity.GoName,
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
		Updatable: updateable(column),
		Null:      nullable(column),
		Min:       0,
		Max:       column.MaxLen,
	}
}

func newSearch(column model.Column, attribute mfd.Attribute, searchType string) *mfd.Search {
	return &mfd.Search{
		Name:       util.ColumnName(mfd.MakeSearchName(attribute.Name, searchType)),
		AttrName:   attribute.Name,
		SearchType: searchType,

		Attribute: &attribute,
	}
}

func PackVTEntity(entity *mfd.Entity) *mfd.VTEntity {
	// processing all columns
	attributes := mfd.VTAttributes{}
	index := mfd.NewSet()

	for _, attr := range entity.Attributes {
		var search *mfd.Search

		// special case for string columns
		if attr.IsString() && !attr.IsArray {
			if search = entity.SearchByAttrName(attr.Name, mfd.SearchILike); search != nil {
				index.Add(search.Name)
			}
		}

		attributes = append(attributes, newVTAttribute(*attr, search))
	}

	// adding searches
	for _, search := range entity.Searches {
		// if was not added already
		if !index.Exists(search.Name) {
			attributes = append(attributes, newVTSearch(*search))
		}
	}

	return &mfd.VTEntity{
		Name:         entity.Name,
		TerminalPath: mfd.UrlName(mfd.MakePlural(entity.Name)),
		Attributes:   attributes,
	}
}

func newVTAttribute(attr mfd.Attribute, search *mfd.Search) *mfd.VTAttribute {
	required := !attr.Nullable()
	if !attr.IsAddable() || !attr.IsUpdatable() {
		required = false
	}

	vtAttr := &mfd.VTAttribute{
		Name:       attr.Name,
		AttrName:   attr.Name,
		SearchName: attr.Name,

		Summary: inSummary(attr),
		Search:  inSearch(attr),

		MaxValue: attr.Max,
		MinValue: attr.Min,
		Required: required,
		Validate: validate(attr),
	}

	// adding search
	if search != nil {
		vtAttr.SearchName = search.Name
		vtAttr.Search = true
	}

	return vtAttr
}

func newVTSearch(search mfd.Search) *mfd.VTAttribute {
	max, min := 0, 0
	if search.Attribute != nil {
		max, min = search.Attribute.Max, search.Attribute.Min
	}

	return &mfd.VTAttribute{
		Name:       search.Name,
		SearchName: search.Name,

		Summary: false,
		Search:  true,

		MaxValue: max,
		MinValue: min,
		Required: false,
	}
}

func PackTemplate(entity *mfd.Entity, vt *mfd.VTEntity) mfd.TmplAttributes {
	tmplAttributes := mfd.TmplAttributes{}

	for _, vtAttr := range vt.Attributes {
		tmpl := &mfd.TmplAttribute{
			Name:     vtAttr.Name,
			AttrName: vtAttr.Name,

			Form:   mfd.TypeHTMLNone,
			Search: mfd.TypeHTMLNone,
		}

		var fk *mfd.TmplAttribute

		attr := entity.AttributeByName(vtAttr.AttrName)
		if attr != nil {
			// not primary key
			if !attr.PrimaryKey && (attr.IsAddable() || attr.IsUpdatable()) {
				tmpl.Form = inputType(*attr, false)
				tmpl.List = inSummary(*attr)
			}

			// foreign key attribute
			if attr.ForeignEntity != nil {
				if title := attr.ForeignEntity.TitleVTAttribute(); title != nil {
					// disable fk in list
					tmpl.List = false
					tmpl.FKOpts = mfd.VarName(title.Name)

					if !attr.IsArray {
						fk = &mfd.TmplAttribute{
							Name:     util.ReplaceSuffix(util.ColumnName(attr.DBName), util.ID, ""),
							AttrName: vtAttr.Name,
							Search:   mfd.TypeHTMLNone,
							List:     true,
							FKOpts:   mfd.VarName(title.Name),
						}
					}
				}
			}
		}

		// adding search
		if search := entity.SearchByName(vtAttr.SearchName); search != nil {
			if mfd.IsArraySearch(search.SearchType) {
				tmpl.Search = mfd.TypeHTMLSelect
			} else {
				tmpl.Search = inputType(*search.Attribute, true)
			}
		} else if searchAttr := entity.AttributeByName(vtAttr.SearchName); searchAttr != nil {
			if !searchAttr.PrimaryKey {
				tmpl.Search = inputType(*searchAttr, true)
			}
		}

		if tmpl.List || tmpl.Form != mfd.TypeHTMLNone || tmpl.Search != mfd.TypeHTMLNone {
			tmplAttributes = append(tmplAttributes, tmpl)
		}

		if fk != nil {
			tmplAttributes = append(tmplAttributes, fk)
		}
	}

	return reorderList(tmplAttributes)
}

func reorderList(attrs mfd.TmplAttributes) mfd.TmplAttributes {
	mp := map[int][]int{}
	for i, attr := range attrs {
		// scoring each column
		score := listScore(attr)
		if _, ok := mp[score]; !ok {
			// storing columns by score
			mp[score] = []int{}
		}

		mp[score] = append(mp[score], i)
	}

	total := 0
	// ranging over all scores
	for score := 1; score <= 6; score++ {
		if _, ok := mp[score]; !ok {
			continue
		}

		// turn off column if limit exceeded
		for _, index := range mp[score] {
			total++
			if total > 7 {
				attrs[index].List = false
			}
		}
	}

	return attrs
}

func listScore(attr *mfd.TmplAttribute) int {
	if !attr.List {
		return -1
	}
	if mfd.IsStatus(attr.Name) {
		return 1
	}
	switch attr.Form {
	case mfd.TypeHTMLCheckbox:
		return 2
	case mfd.TypeHTMLInput:
		return 3
	case mfd.TypeHTMLText:
		return 4
	case mfd.TypeHTMLEditor:
		return 5
	}

	return 6
}

func inputType(attribute mfd.Attribute, forSearch bool) string {
	if attribute.ForeignKey == mfd.VfsFile {
		if strings.Contains(strings.ToLower(attribute.Name), "image") {
			return mfd.TypeHTMLImage
		}
		return mfd.TypeHTMLFile
	}
	if attribute.IsArray && attribute.ForeignKey != "" && !forSearch {
		return mfd.TypeHTMLSelect
	}

	if attribute.IsArray || attribute.IsMap() || attribute.IsJSON() {
		return mfd.TypeHTMLNone
	}

	switch attribute.DBType {
	case model.TypePGText:
		if attribute.Name == "Description" || attribute.Name == "Content" {
			return mfd.TypeHTMLEditor
		}
		return mfd.TypeHTMLText
	case model.TypePGVarchar:
		if attribute.Name == "Password" {
			return mfd.TypeHTMLPassword
		}
		if attribute.Max >= 256 {
			return mfd.TypeHTMLText
		}
		return mfd.TypeHTMLInput
	case model.TypePGBool:
		return mfd.TypeHTMLCheckbox
	case model.TypePGDate:
		return mfd.TypeHTMLDate
	case model.TypePGTime, model.TypePGTimetz:
		return mfd.TypeHTMLTime
	case model.TypePGTimestamp, model.TypePGTimestamptz:
		return mfd.TypeHTMLDateTime
	}

	return mfd.TypeHTMLInput
}

func inSummary(attr mfd.Attribute) bool {
	if attr.Name == "Password" {
		return false
	}

	return !attr.IsArray && !attr.IsJSON() && !attr.IsMap()
}

func inSearch(attr mfd.Attribute) bool {
	if attr.Name == "Password" {
		return false
	}

	return !attr.IsArray && !attr.IsJSON() && !attr.IsMap()
}

func validate(attr mfd.Attribute) string {
	switch attr.DBName {
	case "email", "mail":
		return "email"
	case "ip":
		return "ip"
	case "alias":
		return "alias"
	case "statusId":
		return "status"
	}

	return ""
}

// nullable attribute logic here
func nullable(column model.Column) string {
	switch {
	case column.IsPK || column.Nullable:
		return mfd.NullableYes
	//case column.GoType == model.TypeString || column.IsFK:
	//	return mfd.NullableEmpty
	default:
		return mfd.NullableNo
	}
}

// addable attribute logic here
func addable(column model.Column) *bool {
	result := true
	if column.PGName == "createdAt" || column.PGName == "modifiedAt" {
		result = false
	}

	return &result
}

// updateable attribute logic here
func updateable(column model.Column) *bool {
	result := true
	if column.PGName == "createdAt" || column.PGName == "modifiedAt" {
		result = false
	}

	return &result
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
