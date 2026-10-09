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

func (e EntityData) UsesTableDate() bool {
	for _, column := range e.ListColumns {
		if column.IsTableDate {
			return true
		}
	}

	return false
}

// PackEntity packs mfd vt entity to template data.
// vtTemplate is the normalized project-level template set name
func PackEntity(vtEntity mfd.VTEntity, vtTemplate string) EntityData {
	pks := vtEntity.Entity.PKs()
	pkPairs := make([]PKPair, len(pks))
	for i := range pks {
		pkPairs[i] = PKPair{JSName: mfd.VarName(pks[i].Name)}
	}

	tmpl := EntityData{
		Name:     vtEntity.Name,
		JSName:   mfd.VarName(vtEntity.Name),
		PKs:      pkPairs,
		ReadOnly: vtEntity.Mode == mfd.ModeReadOnlyWithTemplates,
	}

	if title := vtEntity.Entity.TitleAttribute(); title != nil {
		tmpl.HasQuickFilter = true
		tmpl.TitleField = mfd.VarName(title.Name)
	}

	for _, attr := range vtEntity.TmplAttributes {
		if attr.List {
			tmpl.ListColumns = append(tmpl.ListColumns, PackAttribute(vtEntity, *attr))
		}
		if attr.Search != mfd.TypeHTMLNone && attr.Search != "" {
			tmpl.FilterColumns = append(tmpl.FilterColumns, PackInput(*attr, vtEntity, true, vtTemplate))
		}
		if attr.Form != mfd.TypeHTMLNone && attr.Form != "" {
			tmpl.FormColumns = append(tmpl.FormColumns, PackInput(*attr, vtEntity, false, vtTemplate))
		}
	}

	return tmpl
}

const tableDatePipe = "tableDate"

// AttributeData stores attribute info
type AttributeData struct {
	JSName string

	EditLink   bool
	IsBool     bool
	IsSortable bool

	HasPipe bool
	Pipe    template.HTML

	// IsTableDate reports that the column is rendered through the tableDate helper.
	IsTableDate bool

	// Value is the ready to use cell expression for templates without Vue 2
	// filters (vue3): `item.createdAt`, `tableDate(item.createdAt)`, `item.category?.title`
	Value template.HTML
}

// PackAttribute packs mfd tmpl attribute to template data
func PackAttribute(vtEntity mfd.VTEntity, tmpl mfd.TmplAttribute) AttributeData {
	lowerName := strings.ToLower(tmpl.Name)
	boolType := false
	isSortable := true

	jsName := mfd.VarName(tmpl.Name)
	pipe := ""
	value := template.HTML("item." + jsName)
	isTableDate := false

	if tmpl.VTAttribute != nil {
		attr := tmpl.VTAttribute.Attribute

		if attr.IsDateTime() {
			pipe = tableDatePipe
			value = template.HTML(fmt.Sprintf("%s(item.%s)", tableDatePipe, jsName))
			isTableDate = true
		}
		if attr.ForeignKey != "" {
			fkField := mfd.VarName(tmpl.FKOpts)
			pipe = fmt.Sprintf(`getField("%s")`, fkField)
			// vue3 has no filters, so the nested field is read directly
			value = template.HTML(fmt.Sprintf("item.%s?.%s", jsName, fkField))
			isSortable = false
			isTableDate = false
		}
		if attr.IsBool() || tmpl.Search == mfd.TypeHTMLCheckbox {
			boolType = true
			isSortable = false
		}
	}

	return AttributeData{
		JSName:      jsName,
		EditLink:    vtEntity.Mode == mfd.ModeFull && (lowerName == "title" || lowerName == "name"),
		IsBool:      tmpl.List && boolType,
		IsSortable:  isSortable,
		HasPipe:     pipe != "",
		Pipe:        template.HTML(pipe),
		IsTableDate: isTableDate,
		Value:       value,
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

// IsDefaultComponent reports that Component is the fallback text field, so vue3
// templates may skip the explicit component attribute.
func (i InputData) IsDefaultComponent() bool {
	return i.Component == defaultInputComponent
}

// vtDialect holds the js and vuetify api differences between the built-in
// template sets that leak into the packed data.
type vtDialect struct {
	// modelPrefix is how form templates reach the edited model: the mobx store
	// in vue2, the model destructured from useEntityForm in the other sets.
	modelPrefix template.HTML
	// smAndUp is the "small and up" breakpoint expression, renamed in vuetify 3.
	smAndUp template.HTML
}

var vtDialects = map[string]vtDialect{
	mfd.VTTemplateVue2:        {modelPrefix: "store.model.", smAndUp: "$vuetify.breakpoint.smAndUp"},
	mfd.VTTemplateComposition: {modelPrefix: "model.", smAndUp: "$vuetify.breakpoint.smAndUp"},
	mfd.VTTemplateVue3:        {modelPrefix: "model.", smAndUp: "$vuetify.display.smAndUp"},
}

// dialect returns the dialect of a normalized template set name, falling back
// to vue2 the same way mfd.Project.VTTemplateName does.
func dialect(vtTemplate string) vtDialect {
	if d, ok := vtDialects[vtTemplate]; ok {
		return d
	}

	return vtDialects[mfd.VTTemplateVue2]
}

// PackInput packs mfd tmpl attribute to template input data.
// vtTemplate is the normalized project-level template set name, it selects the
// dialect the packed params are rendered in.
func PackInput(tmpl mfd.TmplAttribute, vtEntity mfd.VTEntity, isSearch bool, vtTemplate string) InputData {
	d := dialect(vtTemplate)
	modelPrefix, smAndUp := d.modelPrefix, d.smAndUp

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
			inp.Params = append(inp.Params, `compact`, `:row="`+smAndUp+`"`)
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
			inp.Params = append(inp.Params, `:value-for-transliterating="`+modelPrefix+trasliteratingValue+`"`)
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
			inp.Params = append(inp.Params, `:file="`+modelPrefix+template.HTML(inp.FKJSName)+`"
                    @input:file="file => `+modelPrefix+template.HTML(inp.FKJSName)+` = file"`)
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

// defaultInputComponent is the fallback input component
const defaultInputComponent = "v-text-field"

func filterComponent(input string, isSearch bool) string {
	defaultComponent := defaultInputComponent
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
