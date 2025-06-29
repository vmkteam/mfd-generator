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
		if strings.EqualFold(ns.Name, namespace) {
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

// AddVTEntity adds entity to namespace
func (p *Project) AddVTEntity(namespace string, entity *VTEntity) *VTEntity {
	ns := p.VTNamespace(namespace)
	if ns == nil {
		ns = p.AddVTNamespace(namespace)
	}

	return ns.AddVTEntity(entity)
}

// VTNamespaceNames returns every vt namespaces
func (p *Project) VTNamespaceNames() []string {
	res := make([]string, len(p.VTNamespaces))
	for i := range p.VTNamespaces {
		res[i] = p.VTNamespaces[i].Name
	}

	return res
}

// VTNamespace is xml element
type VTNamespace struct {
	XMLName xml.Name `xml:"VTNamespace" json:"-"`
	XMLxsi  string   `xml:"xmlns:xsi,attr" json:"-"`
	XMLxsd  string   `xml:"xmlns:xsd,attr" json:"-"`
	Name    string

	Entities []*VTEntity `xml:"VTEntities>Entity" json:"vtEntities"`
}

func NewVTNamespace(namespace string) *VTNamespace {
	return &VTNamespace{
		XMLxsi: "",
		XMLxsd: "",

		Name: namespace,

		Entities: []*VTEntity{},
	}
}

// VTEntity returns mfd.VTEntity by its name
func (n *VTNamespace) VTEntity(entity string) *VTEntity {
	for _, e := range n.Entities {
		if strings.EqualFold(e.Name, entity) {
			return e
		}
	}

	return nil
}

// VTEntityIndex returns mfd.VTEntity index by its name
func (n *VTNamespace) VTEntityIndex(entity string) int {
	for i, e := range n.Entities {
		if strings.EqualFold(e.Name, entity) {
			return i
		}
	}

	return -1
}

// AddVTEntity adds vt entity to namespace
func (n *VTNamespace) AddVTEntity(entity *VTEntity) *VTEntity {
	if index := n.VTEntityIndex(entity.Name); index != -1 {
		n.Entities[index] = entity
		return entity
	}

	n.Entities = append(n.Entities, entity)
	return entity
}

// VTEntityNames returns every entity in project.
func (n *VTNamespace) VTEntityNames() []string {
	res := make([]string, len(n.Entities))
	for i := range n.Entities {
		res[i] = n.Entities[i].Name
	}

	return res
}

// VTEntity is xml element
type VTEntity struct {
	XMLName xml.Name `xml:"Entity" json:"-"`
	Name    string   `xml:"Name,attr" json:"name"`

	TerminalPath string `xml:"TerminalPath" json:"terminalPath"`

	Attributes     VTAttributes   `xml:"Attributes>Attribute" json:"attributes"`
	TmplAttributes TmplAttributes `xml:"Template>Attribute" json:"template"`

	// DEPRECATED
	NoTemplates bool `xml:"WithoutTemplates,attr,omitempty" json:"-"`

	Mode string `xml:"Mode,attr" json:"mode"`

	// corresponding entity
	Entity *Entity `xml:"-" json:"-"`
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

// AttributeByNames gets mfd.VTAttribute by its field name
func (e *VTEntity) AttributeByNames(attrName, searchName string) *VTAttribute {
	for _, a := range e.Attributes {
		if a.AttrName == attrName && a.SearchName == searchName {
			return a
		}
	}
	return nil
}

// TmplAttributeByNames gets mfd.TmplAttribute by its field name
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
	Name       string `xml:"Name,attr" json:"name"`
	AttrName   string `xml:"AttrName,attr,omitempty" json:"attrName"`
	SearchName string `xml:"SearchName,attr,omitempty" json:"searchName"`

	Summary bool `xml:"Summary,attr" json:"summary"`
	Search  bool `xml:"Search,attr" json:"search"`

	MaxValue int    `xml:"Max,attr" json:"max"`
	MinValue int    `xml:"Min,attr" json:"min"`
	Required bool   `xml:"Required,attr" json:"required"`
	Validate string `xml:"Validate,attr" json:"validate"`

	// corresponding entity attribute
	Attribute *Attribute `xml:"-" json:"-"`
}

// Merge fills attribute (from db) values from old (in file) attribute
func (a *VTAttribute) Merge(with *VTAttribute) *VTAttribute {
	// a.SearchName = with.SearchName

	if a.Validate == "" {
		a.Validate = with.Validate
	}

	if !a.Required {
		a.Required = with.Required
	}

	return a
}

type TmplAttribute struct {
	XMLName xml.Name `xml:"Attribute" json:"-"`

	Name     string `xml:"Name,attr" json:"name"`
	AttrName string `xml:"VTAttrName,attr" json:"vtAttrName"`

	List   bool   `xml:"List,attr" json:"list"`               // show in list
	FKOpts string `xml:"FKOpts,attr,omitempty" json:"fkOpts"` // how to show fks
	Form   string `xml:"Form,attr" json:"form"`               // show in object editor
	Search string `xml:"Search,attr" json:"search"`           // input type in search

	// corresponding vt attribute
	VTAttribute *VTAttribute `xml:"-" json:"-"`
}

// Merge fills attribute (from db) values from old (in file) attribute
func (a *TmplAttribute) Merge(with *TmplAttribute) {
	if a.FKOpts == "" {
		a.FKOpts = with.FKOpts
	}
}

type VTAttributes []*VTAttribute

// Merge adds new attribute, skip if exists
func (a VTAttributes) Merge(attr *VTAttribute) (VTAttributes, *VTAttribute) {
	for _, existing := range a {
		if (attr.AttrName != "" && existing.AttrName == attr.AttrName) ||
			(attr.AttrName == "" && existing.SearchName == attr.SearchName) {
			return a, existing
		}
	}

	return append(a, attr), attr
}

func (a VTAttributes) updateAttr(entity *Entity) {
	for _, vtAttribute := range a {
		if vtAttribute.AttrName != "" {
			vtAttribute.Attribute = entity.AttributeByName(vtAttribute.AttrName)
		}
		if vtAttribute.SearchName != "" {
			if search := entity.SearchByName(vtAttribute.SearchName); search != nil {
				vtAttribute.Attribute = search.Attribute
			}
		}
	}
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
