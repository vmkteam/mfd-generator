package vttmpl

import (
	"fmt"
	"github.com/vmkteam/mfd-generator/mfd"
	"html/template"
	"strings"
)

const MaxShortFilters = 3

type VTTemplateEntity struct {
	Name   string
	JSName string

	HasQuickFilter bool
	TitleField     string

	PKs []PKPair

	ListColumns   []VTTemplateColumn
	FilterColumns []VTTemplateInput
	FormColumns   []VTTemplateInput
}

func NewVTTemplateEntity(entity mfd.Entity) VTTemplateEntity {
	quickFilter := ""
	if title := entity.TitleVTAttribute(); title != nil {
		quickFilter = mfd.VarName(title.Name)
	}

	var pks []PKPair
	for _, pk := range entity.PKs() {
		pks = append(pks, PKPair{
			JSName: mfd.VarName(pk.Name),
		})
	}

	tmpl := VTTemplateEntity{
		Name:           entity.VTEntity.Name,
		JSName:         mfd.VarName(entity.VTEntity.Name),
		HasQuickFilter: quickFilter != "",
		TitleField:     quickFilter,
		PKs:            pks,
	}

	for _, attr := range entity.VTEntity.TmplAttributes {
		if attr.List {
			tmpl.ListColumns = append(tmpl.ListColumns, NewVTTemplateColumn(*attr, entity))
		}
		if attr.Search != mfd.TypeHTMLNone && attr.Search != "" {
			tmpl.FilterColumns = append(tmpl.FilterColumns, NewVTTemplateInput(*attr, entity, true))
		}
		if attr.Form != mfd.TypeHTMLNone && attr.Form != "" {
			tmpl.FormColumns = append(tmpl.FormColumns, NewVTTemplateInput(*attr, entity, false))
		}
	}

	// making short filters
	for i, filter := range tmpl.FilterColumns {
		if i < MaxShortFilters {
			tmpl.FilterColumns[i].IsShortFilter = true
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

type VTTemplateColumn struct {
	JSName string

	EditLink bool

	HasPipe bool
	Pipe    template.HTML
}

func NewVTTemplateColumn(tmpl mfd.TmplAttribute, entity mfd.Entity) VTTemplateColumn {
	lowerName := strings.ToLower(tmpl.Name)

	pipe := ""
	if vtAttr := entity.VTEntity.Attribute(tmpl.AttrName); vtAttr != nil {
		if attr := entity.AttributeByName(vtAttr.AttrName); attr != nil {
			if attr.IsDateTime() {
				pipe = "tableDate"
			}
			if attr.ForeignKey != "" {
				pipe = fmt.Sprintf(`getField("%s")`, mfd.VarName(tmpl.FKOpts))
			}
		}
	}

	return VTTemplateColumn{
		JSName:   mfd.VarName(tmpl.Name),
		EditLink: lowerName == "title" || lowerName == "name",
		HasPipe:  pipe != "",
		Pipe:     template.HTML(pipe),
	}
}

type VTTemplateInput struct {
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

func NewVTTemplateInput(tmpl mfd.TmplAttribute, entity mfd.Entity, isSearch bool) VTTemplateInput {
	inp := VTTemplateInput{
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

	if vtAttr := entity.VTEntity.Attribute(tmpl.AttrName); vtAttr != nil {
		inp.Required = vtAttr.Required

		if attr := entity.AttributeByName(vtAttr.AttrName); attr != nil && attr.ForeignKey != "" {
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
		return "vt-file-input"
	}

	return "v-text-field"
}
