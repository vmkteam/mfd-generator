package vttmpl

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"
)

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
	SearchType string

	Required bool

	IsArray    bool
	IsCheckBox bool
	IsNumber   bool
	Params     []template.HTML
}

// PackInput packs mfd tmpl attribute to template input data
func PackInput(tmpl mfd.TmplAttribute, vtEntity mfd.VTEntity, isSearch bool) InputData {
	inp := InputData{
		JSName:    mfd.VarName(tmpl.Name),
		Component: filterComponent(tmpl.Search, isSearch),
		Params:    []template.HTML{},
	}

	if !isSearch {
		inp.Component = filterComponent(tmpl.Form, isSearch)
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

	if mfd.MakeJSType(tmpl.VTAttribute.Attribute.GoType, tmpl.VTAttribute.Attribute.IsArray) == "number" {
		inp.IsNumber = true
	}

	if strings.EqualFold(tmpl.Name, "alias") {
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
			inp.Component = filterComponent(tmpl.Form, isSearch)
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
				inp.IsArray = true
			}
		}
	}

	if isSearch {
		inp.SearchType = filterInputType(tmpl.Search, inp)
	}

	return inp
}

func filterComponent(input string, isSearch bool) string {
	defaultComponent := "v-text-field"
	switch input {
	case mfd.TypeHTMLInput:
		return defaultComponent
	case mfd.TypeHTMLCheckbox:
		return "v-checkbox"
	case mfd.TypeHTMLText:
		if isSearch {
			return defaultComponent
		}
		return "v-textarea"
	case mfd.TypeHTMLEditor:
		if isSearch {
			return defaultComponent
		}
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

	return defaultComponent
}

func filterInputType(input string, inp InputData) string {
	if inp.IsFK && inp.IsArray {
		return "multi-select"
	}
	if inp.IsFK || mfd.IsStatus(inp.JSName) {
		return "select"
	}

	switch input {
	case mfd.TypeHTMLInput, mfd.TypeHTMLText, mfd.TypeHTMLEditor:
		return "input"
	case mfd.TypeHTMLDateTime, mfd.TypeHTMLTime:
		return "datetime"
	case mfd.TypeHTMLDate:
		return "date"
	case mfd.TypeHTMLSelect:
		return "select"
	case mfd.TypeHTMLCheckbox:
		return "boolean"
	}

	return "input"
}
