package mfd

import "encoding/xml"

const (
	TypeHTMLNone     = "HTML_NONE"
	TypeHTMLInput    = "HTML_INPUT"
	TypeHTMLText     = "HTML_TEXT"
	TypeHTMLPassword = "HTML_PASSWORD"
	TypeHTMLEditor   = "HTML_EDITOR"
	TypeHTMLCheckbox = "HTML_CHECKBOX"
	TypeHTMLDateTime = "HTML_DATETIME"
	TypeHTMLDate     = "HTML_DATE"
	TypeHTMLTime     = "HTML_TIME"
	TypeHTMLSelect   = "HTML_SELECT"
	TypeHTMLFile     = "HTML_FILE"
)

// VTNamespace is xml element
type VTNamespace struct {
	XMLName xml.Name `xml:"VTNamespace" json:"-"`
	XMLxsi  string   `xml:"xmlns:xsi,attr"`
	XMLxsd  string   `xml:"xmlns:xsd,attr"`

	Entities VTEntities `xml:"VTEntities>Entity"`
}

func NewVTNamespace(entities VTEntities) VTNamespace {
	return VTNamespace{
		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
		XMLxsd: "http://www.w3.org/2001/XMLSchema",

		Entities: entities,
	}
}

// VTEntity is xml element
type VTEntity struct {
	XMLName      xml.Name `xml:"Entity" json:"-"`
	Name         string   `xml:"Name,attr"`
	TerminalPath string   `xml:"TerminalPath"`

	Attributes     VTAttributes   `xml:"Attributes>Attribute"`
	TmplAttributes TmplAttributes `xml:"Template>Attribute"`
}

// Attribute gets mfd.VTAttribute by its field name
func (e *VTEntity) Attribute(name string) *VTAttribute {
	for _, a := range e.Attributes {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// Attribute gets mfd.VTAttribute by its field name
func (e *VTEntity) AttributeByNames(attrName, searchName string) *VTAttribute {
	for _, a := range e.Attributes {
		if a.AttrName == attrName && a.SearchName == searchName {
			return a
		}
	}
	return nil
}

// Attribute gets mfd.TmplAttribute by its field name
func (e *VTEntity) TmplAttributeByNames(name, attrName string) *TmplAttribute {
	for _, a := range e.TmplAttributes {
		if a.Name == name && a.AttrName == attrName {
			return a
		}
	}
	return nil
}

// Merge fills entity (generated from db) with attributes from old (in file) entity
func (e *VTEntity) Merge(with *VTEntity) {
	attrs := e.Attributes
	for _, toAdd := range with.Attributes {
		// adding only new
		if existing := e.AttributeByNames(toAdd.AttrName, toAdd.SearchName); existing != nil {
			existing.Merge(toAdd)
		} else {
			attrs = append(attrs, toAdd)
		}
	}

	e.Attributes = attrs
}

func (e *VTEntity) AddTmpl(attrs TmplAttributes) {
	tmplAttrs := e.TmplAttributes
	for _, toAdd := range attrs {
		// adding only new
		if existing := e.TmplAttributeByNames(toAdd.Name, toAdd.AttrName); existing != nil {
			existing.Merge(toAdd)
		} else {
			tmplAttrs = append(tmplAttrs, toAdd)
		}
	}

	e.TmplAttributes = tmplAttrs
}

// VTAttribute is xml element
type VTAttribute struct {
	XMLName xml.Name `xml:"Attribute" json:"-"`

	// Names
	Name       string `xml:"Name,attr"`
	AttrName   string `xml:"AttrName,attr,omitempty"`
	SearchName string `xml:"SearchName,attr,omitempty"`

	// model options
	Summary bool `xml:"Summary,attr"` // show in list
	Search  bool `xml:"Search,attr"`  // show in search

	// Validate options
	MaxValue int    `xml:"Max,attr"`
	MinValue int    `xml:"Min,attr"`
	Required bool   `xml:"Required,attr"`
	Validate string `xml:"Validate,attr"`
}

// Merge fills attribute (from db) values from old (in file) attribute
func (a *VTAttribute) Merge(with *VTAttribute) {
	a.SearchName = with.SearchName

	if a.Validate == "" {
		a.Validate = with.Validate
	}

	if !a.Required {
		a.Required = with.Required
	}
}

type TmplAttribute struct {
	XMLName xml.Name `xml:"Attribute" json:"-"`

	Name     string `xml:"Name,attr"`
	AttrName string `xml:"VTAttrName,attr"`

	List   bool   `xml:"List,attr"`             // show in list
	FKOpts string `xml:"FKOpts,attr,omitempty"` // how to show fks
	Form   string `xml:"Form,attr"`             // show in object editor
	Search string `xml:"Search,attr"`           // input type in search
}

// Merge fills attribute (from db) values from old (in file) attribute
func (a *TmplAttribute) Merge(with *TmplAttribute) {
	if a.FKOpts == "" {
		a.FKOpts = with.FKOpts
	}
}

// convenient types
type VTEntities []*VTEntity

type VTAttributes []*VTAttribute

type TmplAttributes []*TmplAttribute
