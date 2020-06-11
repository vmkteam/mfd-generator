package vttmpl

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"
)

const MaxShortFilters = 3

// EntityData stores entity info
type EntityData struct {
	Name   string
	JSName string

	HasQuickFilter bool
	TitleField     string

	PKs []PKPair

	ReadOnly bool

	ListColumns   []AttributeData
	FilterColumns []InputData
	FormColumns   []InputData
}

// PackEntity packs mfd vt entity to template data
func PackEntity(vtEntity mfd.VTEntity) EntityData {
	var pks []PKPair
	for _, pk := range vtEntity.Entity.PKs() {
		pks = append(pks, PKPair{
			JSName: mfd.VarName(pk.Name),
		})
	}
	quickFilter := ""
	if title := vtEntity.Entity.TitleAttribute(); title != nil {
		quickFilter = mfd.VarName(title.Name)
	}

	tmpl := EntityData{
		Name:           vtEntity.Name,
		JSName:         mfd.VarName(vtEntity.Name),
		HasQuickFilter: quickFilter != "", // TODO remove
		TitleField:     quickFilter,       // TODO remove
		PKs:            pks,
		ReadOnly:       vtEntity.Mode == mfd.ModeReadOnlyWithTemplates,
	}

	for _, attr := range vtEntity.TmplAttributes {
		if attr.List {
			tmpl.ListColumns = append(tmpl.ListColumns, PackAttribute(vtEntity, *attr))
		}
		if attr.Search != mfd.TypeHTMLNone && attr.Search != "" {
			tmpl.FilterColumns = append(tmpl.FilterColumns, PackInput(*attr, vtEntity, true))
		}
		if attr.Form != mfd.TypeHTMLNone && attr.Form != "" {
			tmpl.FormColumns = append(tmpl.FormColumns, PackInput(*attr, vtEntity, false))
		}
	}

	// making short filters
	for i, filter := range tmpl.FilterColumns {
		if i < MaxShortFilters {
			tmpl.FilterColumns[i].IsShortFilter = true
		}

		if i == MaxShortFilters {
			tmpl.FilterColumns[i-1].ShowShortFilterLabel = true
		}

		// status will override last short filter
		if i >= MaxShortFilters && mfd.IsStatus(filter.JSName) {
			tmpl.FilterColumns[MaxShortFilters-1].IsShortFilter = false
			tmpl.FilterColumns[MaxShortFilters-1].ShowShortFilterLabel = false

			tmpl.FilterColumns[i].IsShortFilter = true
			tmpl.FilterColumns[i].ShowShortFilterLabel = true
		}
	}

	return tmpl
}

// AttributeData stores attribute info
type AttributeData struct {
	JSName string

	EditLink   bool
	IsBool     bool
	IsSortable bool

	HasPipe bool
	Pipe    template.HTML
}

// PackAttribute packs mfd tmpl attribute to template data
func PackAttribute(vtEntity mfd.VTEntity, tmpl mfd.TmplAttribute) AttributeData {
	lowerName := strings.ToLower(tmpl.Name)
	boolType := false
	isSortable := true

	pipe := ""
	if tmpl.VTAttribute != nil {
		attr := tmpl.VTAttribute.Attribute

		if attr.IsDateTime() {
			pipe = "tableDate"
		}
		if attr.ForeignKey != "" {
			pipe = fmt.Sprintf(`getField("%s")`, mfd.VarName(tmpl.FKOpts))
			isSortable = false
		}
		if attr.IsBool() || tmpl.Search == mfd.TypeHTMLCheckbox {
			boolType = true
			isSortable = false
		}
	}

	return AttributeData{
		JSName:     mfd.VarName(tmpl.Name),
		EditLink:   vtEntity.Mode == mfd.ModeFull && (lowerName == "title" || lowerName == "name"),
		IsBool:     tmpl.List && boolType,
		IsSortable: isSortable,
		HasPipe:    pipe != "",
		Pipe:       template.HTML(pipe),
	}
}

// InputData stores attribute info for inputs
type InputData struct {
	JSName string

	Component  string
	IsFK       bool
	FKJSName   string
	FKJSSearch string

	IsShortFilter        bool
	ShowShortFilterLabel bool

	Required bool

	IsCheckBox bool
	Params     []template.HTML
}

// PackInput packs mfd tmpl attribute to template input data
func PackInput(tmpl mfd.TmplAttribute, vtEntity mfd.VTEntity, isSearch bool) InputData {
	inp := InputData{
		JSName:    mfd.VarName(tmpl.Name),
		Component: filterComponent(tmpl.Search),
		Params:    []template.HTML{},
	}

	if !isSearch {
		inp.Component = filterComponent(tmpl.Form)
	}

	if mfd.IsStatus(tmpl.Name) {
		inp.Component = "vt-status-select"

		if !isSearch {
			inp.Params = append(inp.Params, `compact`, `:row="$vuetify.breakpoint.smAndUp"`)
		}
	}

	if tmpl.Form == mfd.TypeHTMLPassword && !isSearch {
		inp.Params = append(inp.Params, `type="password"`)
	}

	if tmpl.Form == mfd.TypeHTMLEditor && !isSearch {
		inp.Params = append(inp.Params, `without-help`)
	}

	if tmpl.Form == mfd.TypeHTMLCheckbox || tmpl.Search == mfd.TypeHTMLCheckbox {
		inp.IsCheckBox = true
	}

	if strings.ToLower(tmpl.Name) == "alias" {
		if title := vtEntity.Entity.TitleAttribute(); title != nil {
			trasliteratingValue := template.HTML(mfd.VarName(title.Name))

			inp.Component = "vt-transliterator"
			inp.Params = append(inp.Params, `:value-for-transliterating="store.model.`+trasliteratingValue+`"`)
		}
	}

	if tmpl.VTAttribute != nil {
		inp.Required = tmpl.VTAttribute.Required

		attr := tmpl.VTAttribute.Attribute

		if attr.ForeignKey == mfd.VfsFile {
			inp.Component = filterComponent(tmpl.Form)
			inp.IsFK = false
			inp.FKJSName = mfd.VarName(mfd.FKName(tmpl.AttrName))
			inp.FKJSSearch = mfd.VarName(tmpl.FKOpts)
			inp.Params = append(inp.Params, `:file="store.model.`+template.HTML(inp.FKJSName)+`"
                    @input:file="file => store.model.`+template.HTML(inp.FKJSName)+` = file"`)
		} else if attr.ForeignKey != "" {
			inp.Component = "vt-entity-autocomplete"
			inp.IsFK = true
			inp.FKJSName = mfd.VarName(attr.ForeignEntity.Name)
			inp.FKJSSearch = mfd.VarName(tmpl.FKOpts)
			if attr.IsArray {
				inp.Params = append(inp.Params, `multiple`, `chips`)
			}
		}
	}

	return inp
}

func filterComponent(input string) string {
	switch input {
	case mfd.TypeHTMLInput:
		return "v-text-field"
	case mfd.TypeHTMLCheckbox:
		return "v-checkbox"
	case mfd.TypeHTMLText:
		return "v-textarea"
	case mfd.TypeHTMLEditor:
		return "vt-tinymce-editor"
	case mfd.TypeHTMLDateTime:
		return "vt-datetime-picker"
	case mfd.TypeHTMLTime:
		return "vt-time-picker"
	case mfd.TypeHTMLDate:
		return "vt-date-picker"
	case mfd.TypeHTMLFile:
		return "vt-vfs-file-input"
	case mfd.TypeHTMLImage:
		return "vt-vfs-image-input"
	}

	return "v-text-field"
}
