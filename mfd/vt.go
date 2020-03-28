package mfd

import (
	"encoding/xml"
	"strings"
)

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
	TypeHTMLImage    = "HTML_IMAGE"
)

const (
	ModeFull                  = "Full"
	ModeNone                  = "None"
	ModeReadOnly              = "ReadOnly"
	ModeReadOnlyWithTemplates = "ReadOnlyWithTemplates"
)

// VTNamespace returns mfd.VTNamespace by its name
func (p *Project) VTNamespace(namespace string) *VTNamespace {
	for _, ns := range p.VTNamespaces {
		if strings.ToLower(ns.Name) == strings.ToLower(namespace) {
			return ns
		}
	}

	return nil
}

// AddVTNamespace adds namespace and return it
func (p *Project) AddVTNamespace(namespace string) *VTNamespace {
	ns := NewVTNamespace(namespace)

	p.NamespaceNames = append(p.NamespaceNames, namespace)
	p.VTNamespaces = append(p.VTNamespaces, ns)

	return ns
}

// AddEntity adds entity to namespace
func (p *Project) AddVTEntity(namespace string, entity *VTEntity) *VTEntity {
	ns := p.VTNamespace(namespace)
	if ns == nil {
		ns = p.AddVTNamespace(namespace)
	}

	return ns.AddVTEntity(entity)
}

// VTNamespace is xml element
type VTNamespace struct {
	XMLName xml.Name `xml:"VTNamespace" json:"-"`
	XMLxsi  string   `xml:"xmlns:xsi,attr"`
	XMLxsd  string   `xml:"xmlns:xsd,attr"`
	Name    string

	Entities []*VTEntity `xml:"VTEntities>Entity"`
}

func NewVTNamespace(namespace string) *VTNamespace {
	return &VTNamespace{
		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
		XMLxsd: "http://www.w3.org/2001/XMLSchema",

		Name: namespace,

		Entities: []*VTEntity{},
	}
}

// VTEntity returns mfd.VTEntity by its name
func (n *VTNamespace) VTEntity(entity string) *VTEntity {
	for _, e := range n.Entities {
		if strings.ToLower(e.Name) == strings.ToLower(entity) {
			return e
		}
	}

	return nil
}

// VTEntityIndex returns mfd.VTEntity index by its name
func (n *VTNamespace) VTEntityIndex(entity string) int {
	for i, e := range n.Entities {
		if strings.ToLower(e.Name) == strings.ToLower(entity) {
			return i
		}
	}

	return -1
}

// AddEntity adds entity to namespace
func (n *VTNamespace) AddVTEntity(entity *VTEntity) *VTEntity {
	if index := n.VTEntityIndex(entity.Name); index != -1 {
		n.Entities[index] = entity
		return entity
	}

	n.Entities = append(n.Entities, entity)
	return entity
}

// VTEntity is xml element
type VTEntity struct {
	XMLName xml.Name `xml:"Entity" json:"-"`
	Name    string   `xml:"Name,attr"`

	TerminalPath string `xml:"TerminalPath"`

	Attributes     VTAttributes   `xml:"Attributes>Attribute"`
	TmplAttributes TmplAttributes `xml:"Template>Attribute"`

	// DEPRECATED
	NoTemplates bool   `xml:"WithoutTemplates,attr,omitempty"`
	Mode        string `xml:"Mode,attr"`

	// corresponding entity
	Entity *Entity `xml:"-"`
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

	// corresponding entity attribute
	Attribute *Attribute `xml:"-"`
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

	// corresponding vt attribute
	VTAttribute *VTAttribute `xml:"-"`
}

// Merge fills attribute (from db) values from old (in file) attribute
func (a *TmplAttribute) Merge(with *TmplAttribute) {
	if a.FKOpts == "" {
		a.FKOpts = with.FKOpts
	}
}

type VTAttributes []*VTAttribute

// Merge adds new attribute, update if exists
func (a VTAttributes) Merge(attr *VTAttribute) (VTAttributes, *VTAttribute) {
	for i, existing := range a {
		if existing.AttrName == attr.AttrName && existing.SearchName == attr.SearchName {
			existing.Merge(attr)
			a[i] = existing
			return a, existing
		}
	}

	return append(a, attr), attr
}

type TmplAttributes []*TmplAttribute

// Merge adds new attribute, update if exists
func (a TmplAttributes) Merge(tmpl *TmplAttribute) (TmplAttributes, *TmplAttribute) {
	for i, existing := range a {
		if existing.AttrName == tmpl.AttrName && existing.Name == tmpl.Name {
			existing.Merge(tmpl)
			a[i] = existing
			return a, existing
		}
	}

	return append(a, tmpl), tmpl
}
