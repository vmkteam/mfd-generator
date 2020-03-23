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

// Entity returns mfd.Entity by its name
func (n *VTNamespace) VTEntity(entity string) *VTEntity {
	for _, e := range n.Entities {
		if strings.ToLower(e.Name) == strings.ToLower(entity) {
			return e
		}
	}

	return nil
}

// AddEntity adds entity to namespace
func (n *VTNamespace) AddVTEntity(entity *VTEntity) *VTEntity {
	if existing := n.VTEntity(entity.Name); existing != nil {
		existing.Merge(entity)
		return existing
	}

	n.Entities = append(n.Entities, entity)
	return entity
}

// VTEntity is xml element
type VTEntity struct {
	XMLName     xml.Name `xml:"Entity" json:"-"`
	Name        string   `xml:"Name,attr"`
	NoTemplates bool     `xml:"WithoutTemplates,attr"`

	TerminalPath string `xml:"TerminalPath"`

	Attributes     []*VTAttribute   `xml:"Attributes>Attribute"`
	TmplAttributes []*TmplAttribute `xml:"Template>Attribute"`

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

func (e *VTEntity) AddTmpl(attrs []*TmplAttribute) {
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
