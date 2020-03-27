package mfd

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

// go-pg versions
const (
	GoPG8 = 8
	GoPG9 = 9
)

// nullable options
const (
	NullableYes   = "Yes"
	NullableNo    = "No"
	NullableEmpty = "CheckEmpty"
)

// vfsFile entity name
const (
	VfsFile = "VfsFile"
)

// TODO Refactor
// search types
const (
	SearchEquals            = "SEARCHTYPE_EQUALS"
	SearchNotEquals         = "SEARCHTYPE_NOT_EQUALS"
	SearchNull              = "SEARCHTYPE_NULL"
	SearchNotNull           = "SEARCHTYPE_NOT_NULL"
	SearchGE                = "SEARCHTYPE_GE"
	SearchLE                = "SEARCHTYPE_LE"
	SearchG                 = "SEARCHTYPE_G"
	SearchL                 = "SEARCHTYPE_L"
	SearchLeftLike          = "SEARCHTYPE_LEFT_LIKE"
	SearchLeftILike         = "SEARCHTYPE_LEFT_ILIKE"
	SearchRightLike         = "SEARCHTYPE_RIGHT_LIKE"
	SearchRightILike        = "SEARCHTYPE_RIGHT_ILIKE"
	SearchLike              = "SEARCHTYPE_LIKE"
	SearchILike             = "SEARCHTYPE_ILIKE"
	SearchArray             = "SEARCHTYPE_ARRAY"
	SearchNotArray          = "SEARCHTYPE_NOT_INARRAY"
	SearchTypeArrayContains = "SEARCHTYPE_ARRAY_CONTAINS"
)

// Project is xml element
type Project struct {
	XMLName        xml.Name `xml:"Project" json:"-"`
	XMLxsi         string   `xml:"xmlns:xsi,attr"`
	XMLxsd         string   `xml:"xmlns:xsd,attr"`
	Name           string
	NamespaceNames []string `xml:"PackageNames>string" json:"-"`
	Languages      []string `xml:"Languages>string" json:"-"`
	GoPGVer        int      `xml:"GoPGVer"`

	Namespaces   []*Namespace   `xml:"-"`
	VTNamespaces []*VTNamespace `xml:"-"`
}

func NewProject(name string) *Project {
	return &Project{
		Name:           name,
		NamespaceNames: []string{},

		GoPGVer: GoPG8,

		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
		XMLxsd: "http://www.w3.org/2001/XMLSchema",
	}
}

// Namespace returns mfd.Namespace by its name
func (p *Project) Namespace(namespace string) *Namespace {
	for _, ns := range p.Namespaces {
		if strings.ToLower(ns.Name) == strings.ToLower(namespace) {
			return ns
		}
	}

	return nil
}

// AddNamespace adds namespace and return it
func (p *Project) AddNamespace(namespace string) *Namespace {
	ns := NewNamespace(namespace)

	p.NamespaceNames = append(p.NamespaceNames, namespace)
	p.Namespaces = append(p.Namespaces, ns)

	return ns
}

// Entity returns mfd.Entity by its name
func (p *Project) Entity(entity string) *Entity {
	for _, n := range p.Namespaces {
		if e := n.Entity(entity); e != nil {
			return e
		}
	}
	return nil
}

// Entity returns mfd.VTEntity by its name
func (p *Project) VTEntity(entity string) *VTEntity {
	for _, n := range p.VTNamespaces {
		if e := n.VTEntity(entity); e != nil {
			return e
		}
	}
	return nil
}

// AddEntity adds entity to namespace
func (p *Project) AddEntity(namespace string, entity *Entity) *Entity {
	ns := p.Namespace(namespace)
	if ns == nil {
		ns = p.AddNamespace(namespace)
	}

	return ns.AddEntity(entity)
}

func (p *Project) IsConsistent() error {
	if p.GoPGVer != GoPG9 && p.GoPGVer != GoPG8 {
		return fmt.Errorf("unsupported go-pg version: %d", p.GoPGVer)
	}

	for _, nsName := range p.NamespaceNames {
		ns := p.Namespace(nsName)
		if ns == nil {
			return fmt.Errorf("namespace %s listed in names but not found", nsName)
		}

		for _, entity := range ns.Entities {
			for _, attr := range entity.Attributes {
				if attr.ForeignKey != "" && attr.ForeignEntity == nil {
					return fmt.Errorf("fk entity %s not found for %s column in %s entity %s namespace", attr.ForeignKey, attr.Name, entity.Name, nsName)
				}
			}

			for _, search := range entity.Searches {
				if search.Attribute == nil || search.Entity == nil {
					return fmt.Errorf("attribute %s not found for %s search in %s entity %s namespace", search.AttrName, search.Name, entity.Name, nsName)
				}
			}
		}
	}

	for _, vtNamespace := range p.VTNamespaces {
		ns := p.Namespace(vtNamespace.Name)
		if ns == nil {
			return fmt.Errorf("namespace %s not found for vt", vtNamespace.Name)
		}

		for _, vtEntity := range vtNamespace.Entities {
			if vtEntity.Entity == nil {
				return fmt.Errorf("entity not found vtEntity %s in %s namespace", vtEntity.Name, vtNamespace.Name)
			}

			for _, vtAttribute := range vtEntity.Attributes {
				if vtAttribute.Attribute == nil {
					if vtAttribute.AttrName != "" {
						return fmt.Errorf("attribute %s not found for attribute %s in vtEntity %s in %s namespace", vtAttribute.AttrName, vtAttribute.Name, vtEntity.Name, vtNamespace.Name)
					}
					if vtAttribute.SearchName != "" {
						return fmt.Errorf("search %s not found for attribute %s in vtEntity %s in %s namespace", vtAttribute.SearchName, vtAttribute.Name, vtEntity.Name, vtNamespace.Name)
					}
				}
			}
		}
	}

	return nil
}

func (p *Project) SuggestArrayLinks() {
	for _, namespace := range p.Namespaces {
		for _, entity := range namespace.Entities {
			for _, attr := range entity.Attributes {
				// skipping not fks and not arrays
				if !attr.IsForeignKey() || !attr.IsIDsArray() {
					continue
				}

				entityName := strings.TrimSuffix(attr.Name, util.IDs)
				if foreign := p.Entity(entityName); foreign != nil {
					attr.ForeignEntity = foreign
					attr.ForeignKey = foreign.Name
				}
			}
		}
	}
}

func (p *Project) UpdateLinks() {
	// making links
	for _, namespace := range p.Namespaces {
		for _, entity := range namespace.Entities {
			// making fk links
			for _, attr := range entity.Attributes {
				if foreign := p.Entity(attr.ForeignKey); foreign != nil {
					attr.ForeignEntity = foreign
				}
			}

			// making search links
			for _, search := range entity.Searches {
				// attach own attribute and entity
				if attr := entity.AttributeByName(search.AttrName); attr != nil {
					search.Attribute = attr
					search.Entity = entity
				}

				if search.IsForeignSearch() {
					foreignName, foreignAttribute := search.ForeignAttribute()

					if foreign := p.Entity(foreignName); foreign != nil {
						if attr := foreign.AttributeByName(foreignAttribute); attr != nil {
							search.Attribute = attr
							search.Entity = foreign
						}
					}
				}
			}
		}
	}

	// making vt links
	for _, vtNamespace := range p.VTNamespaces {
		for _, vtEntity := range vtNamespace.Entities {
			if entity := p.Entity(vtEntity.Name); entity != nil {
				vtEntity.Entity = entity
				for _, vtAttribute := range vtEntity.Attributes {
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

			for _, tmpl := range vtEntity.TmplAttributes {
				tmpl.VTAttribute = vtEntity.Attribute(tmpl.AttrName)
			}
		}
	}
}

// Namespace is xml element
type Namespace struct {
	XMLName xml.Name `xml:"Package" json:"-"`
	XMLxsi  string   `xml:"xmlns:xsi,attr"`
	XMLxsd  string   `xml:"xmlns:xsd,attr"`
	Name    string

	Entities []*Entity `xml:"Entities>Entity"`
}

// NewNamespace creates Namespace
func NewNamespace(name string) *Namespace {
	return &Namespace{
		Name: name,

		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
		XMLxsd: "http://www.w3.org/2001/XMLSchema",
	}
}

// Entity returns mfd.Entity by its name
func (n *Namespace) Entity(entity string) *Entity {
	for _, e := range n.Entities {
		if strings.ToLower(e.Name) == strings.ToLower(entity) {
			return e
		}
	}

	return nil
}

// Entity returns mfd.Entity index by its name
func (n *Namespace) EntityIndex(entity string) int {
	for i, e := range n.Entities {
		if strings.ToLower(e.Name) == strings.ToLower(entity) {
			return i
		}
	}

	return -1
}

// AddEntity adds entity to namespace
func (n *Namespace) AddEntity(entity *Entity) *Entity {
	if index := n.EntityIndex(entity.Name); index != -1 {
		n.Entities[index] = entity
		return entity
	}

	n.Entities = append(n.Entities, entity)
	return entity
}

// Entity is xml element
type Entity struct {
	XMLName xml.Name `xml:"Entity" json:"-"`

	Name      string `xml:"Name,attr"`
	Namespace string `xml:"Namespace,attr"`
	Table     string `xml:"Table,attr"`

	Attributes Attributes `xml:"Attributes>Attribute,omitempty"`
	Searches   Searches   `xml:"Searches>Search,omitempty"`
}

// AttributeByName gets mfd.Attribute by its name
func (e *Entity) AttributeByName(name string) *Attribute {
	for _, a := range e.Attributes {
		if a.Name == name {
			return a
		}
	}

	return nil
}

// AttributeByDBName gets mfd.Attribute by its db name and type
func (e *Entity) AttributeByDBName(dbName, dbType string) *Attribute {
	for _, a := range e.Attributes {
		if a.DBName == dbName && a.DBType == dbType {
			return a
		}
	}

	return nil
}

// SearchByName gets mfd.Search by its name
func (e *Entity) SearchByName(name string) *Search {
	for _, s := range e.Searches {
		if s.Name == name {
			return s
		}
	}

	return nil
}

// SearchByAttrName gets mfd.Search by its attribute and searchType
func (e *Entity) SearchByAttrName(attrName, searchType string) *Search {
	for _, s := range e.Searches {
		if s.AttrName == attrName && s.SearchType == searchType {
			return s
		}
	}

	return nil
}

// HasMultiplePKs returns true if mfd.Entity has several PKs
func (e *Entity) HasMultiplePKs() bool {
	count := 0
	for _, a := range e.Attributes {
		if a.PrimaryKey {
			count++
		}
	}
	return count > 1
}

// PKs returns PKs for entity
func (e *Entity) PKs() Attributes {
	var pks Attributes
	for _, a := range e.Attributes {
		if a.PrimaryKey {
			pks = append(pks, a)
		}
	}

	return pks
}

func (e *Entity) TitleAttribute() *Attribute {
	for _, attr := range e.Attributes {
		if attr.Name == "Title" || attr.Name == "Name" || attr.Name == "Login" || attr.Name == "Alias" {
			return attr
		}
	}

	if pks := e.PKs(); len(pks) > 0 {
		return pks[0]
	}

	return nil
}

// Attribute is xml element
type Attribute struct {
	XMLName xml.Name `xml:"Attribute" json:"-"`
	// names
	Name   string `xml:"Name,attr"`
	DBName string `xml:"DBName,attr"`

	// types
	IsArray bool   `xml:"IsArray,attr,omitempty"`
	DBType  string `xml:"DBType,attr,omitempty"`
	GoType  string `xml:"GoType,attr,omitempty"`

	// Keys
	PrimaryKey    bool    `xml:"PK,attr"`
	ForeignKey    string  `xml:"FK,attr,omitempty"`
	ForeignEntity *Entity `xml:"-"`

	// data params
	Null      string `xml:"Nullable,attr"`
	Addable   *bool  `xml:"Addable,attr"`
	Updatable *bool  `xml:"Updatable,attr"`
	Min       int    `xml:"Min,attr"`
	Max       int    `xml:"Max,attr"`
	Default   string `xml:"Default,attr,omitempty"`
}

// Merge fills attribute (from file) values from db
func (a *Attribute) Merge(with *Attribute) {
	// a.Name = with.Name
	a.DBName = with.DBName
	a.DBType = with.DBType
	a.IsArray = with.IsArray
	a.ForeignKey = with.ForeignKey
	a.ForeignEntity = with.ForeignEntity

	if a.Addable == nil {
		addable := true
		a.Addable = &addable
	}

	if a.Updatable == nil {
		updatable := true
		a.Updatable = &updatable
	}
}

func (a *Attribute) IsInteger() bool {
	return a.DBType == model.TypePGInt2 || a.DBType == model.TypePGInt4 || a.DBType == model.TypePGInt8
}

func (a *Attribute) IsString() bool {
	return a.DBType == model.TypePGText || a.DBType == model.TypePGVarchar
}

func (a *Attribute) IsBool() bool {
	return a.DBType == model.TypePGBool
}

func (a *Attribute) IsDateTime() bool {
	return a.DBType == model.TypePGTimestamp || a.DBType == model.TypePGTimestamptz ||
		a.DBType == model.TypePGTime || a.DBType == model.TypePGTimetz || a.DBType == model.TypePGDate
}

func (a *Attribute) IsJSON() bool {
	return a.DBType == model.TypePGJSONB || a.DBType == model.TypePGJSON
}

func (a *Attribute) IsMap() bool {
	return a.GoType == model.TypeMapInterface || a.GoType == model.TypeMapString
}

func (a *Attribute) Nullable() bool {
	return a.Null == NullableYes
}

func (a *Attribute) IsAddable() bool {
	return a.Addable == nil || *a.Addable
}

func (a *Attribute) IsUpdatable() bool {
	return a.Updatable == nil || *a.Updatable
}

func (a *Attribute) IsForeignKey() bool {
	return a.ForeignKey == ""
}

func (a *Attribute) IsIDsArray() bool {
	return strings.HasSuffix(a.Name, util.IDs) && a.IsArray
}

// Attribute is xml element
type Search struct {
	XMLName xml.Name `xml:"Search" json:"-"`
	// names
	Name       string `xml:"Name,attr"`
	AttrName   string `xml:"AttrName,attr"`
	SearchType string `xml:"SearchType,attr"`

	Attribute *Attribute `xml:"-"`
	Entity    *Entity    `xml:"-"`
}

func (s *Search) IsForeignSearch() bool {
	return strings.Index(s.AttrName, ".") != -1
}

func (s *Search) ForeignAttribute() (entity, attribute string) {
	parts := strings.SplitN(s.AttrName, ".", 2)
	return util.EntityName(parts[0]), util.ColumnName(parts[1])
}

func IsStatus(name string) bool {
	return strings.ToLower(name) == "statusid" || strings.ToLower(name) == "status"
}

func IsArraySearch(search string) bool {
	return search == SearchArray || search == SearchNotArray
}

type Searches []*Search

// Append adds search to collection if not exists
func (s Searches) Append(search *Search) Searches {
	for _, existing := range s {
		if existing.AttrName == search.AttrName && existing.SearchType == search.SearchType {
			return s
		}
	}

	return append(s, search)
}

type Attributes []*Attribute

// Merge adds new attribute, update if exists
func (a Attributes) Merge(attr *Attribute) (Attributes, *Attribute) {
	for i, existing := range a {
		if existing.DBName == attr.DBName && existing.DBType == attr.DBType {
			existing.Merge(attr)
			a[i] = existing
			return a, existing
		}
	}

	return append(a, attr), attr
}
