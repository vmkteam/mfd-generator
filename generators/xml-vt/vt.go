package xmlvt

import (
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

var (
	AttrPassword = "Password"
)

// this code used to convert entities from database to namespace in mfd project file

func PackVTEntity(entity *mfd.Entity, existing *mfd.VTEntity) *mfd.VTEntity {
	// making copy
	vtEntity := mfd.VTEntity{
		Name:         entity.Name,
		TerminalPath: mfd.URLName(mfd.MakePlural(entity.Name)),
		Attributes:   mfd.VTAttributes{},
		Mode:         mfd.ModeFull,
	}

	if existing != nil {
		vtEntity = *existing
	}

	index := mfd.NewSet()

	for _, attr := range entity.Attributes {
		var search *mfd.Search

		// special case for string columns
		if attr.IsString() && !attr.IsArray {
			if search = entity.SearchByAttrName(attr.Name, mfd.SearchILike); search != nil {
				index.Add(search.Name)
			}
		}

		vtEntity.Attributes, _ = vtEntity.Attributes.Merge(newVTAttribute(*attr, search))
	}

	// adding searches
	for _, search := range entity.Searches {
		// if was not added already
		if !index.Exists(search.Name) {
			vtEntity.Attributes, _ = vtEntity.Attributes.Merge(newVTSearch(*search))
		}
	}

	// adding template
	vtEntity.TmplAttributes = PackTemplate(entity, &vtEntity, existing)

	return &vtEntity
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
	res := &mfd.VTAttribute{
		Name:       search.Name,
		SearchName: search.Name,

		Summary:  false,
		Search:   true,
		Required: false,
	}

	if search.Attribute != nil {
		res.MaxValue = search.Attribute.Max
		res.MinValue = search.Attribute.Min
	}

	return res
}

func PackTemplate(entity *mfd.Entity, vt *mfd.VTEntity, existing *mfd.VTEntity) mfd.TmplAttributes {
	var tmplAttributes mfd.TmplAttributes
	if existing != nil {
		tmplAttributes = existing.TmplAttributes
	}

	for _, vtAttr := range vt.Attributes {
		// if vtAttribute already exists in file - do not generate tmpl for it
		if existing != nil && existing.Attribute(vtAttr.Name) != nil {
			continue
		}

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
				if title := attr.ForeignEntity.TitleAttribute(); title != nil {
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
			if search.SearchType.IsArraySearch() {
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
			tmplAttributes, _ = tmplAttributes.Merge(tmpl)
		}

		if fk != nil {
			tmplAttributes, _ = tmplAttributes.Merge(fk)
		}
	}

	// sorting only if new
	if existing == nil {
		reorderList(tmplAttributes)
	}

	return tmplAttributes
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
		if attribute.Name == AttrPassword {
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
	if attr.Name == AttrPassword {
		return false
	}

	return !attr.IsArray && !attr.IsJSON() && !attr.IsMap()
}

func inSearch(attr mfd.Attribute) bool {
	if attr.Name == AttrPassword {
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
