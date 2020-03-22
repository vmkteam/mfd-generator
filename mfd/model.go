package mfd

import (
	"encoding/xml"
	"strings"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"

	"golang.org/x/xerrors"
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
	NamespaceNames []string   `xml:"PackageNames>string" json:"-"`
	Namespaces     Namespaces `xml:"-"`
	Filename       string     `xml:"-"`
}

func NewProject(name string) *Project {
	return &Project{
		Name:           name,
		NamespaceNames: []string{},

		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
		XMLxsd: "http://www.w3.org/2001/XMLSchema",
	}
}

func (p *Project) IsConsistent() error {
	for _, nsName := range p.NamespaceNames {
		ns := p.Namespace(nsName)
		if ns == nil {
			return xerrors.Errorf("namespace %s listed in names but not found", nsName)
		}

		for _, entity := range ns.Entities {
			for _, attr := range entity.Attributes {
				if attr.ForeignKey != "" && attr.ForeignEntity == nil {
					return xerrors.Errorf("fk entity %s not found for %s column in %s entity %s namespace", attr.ForeignKey, attr.Name, entity.Name, nsName)
				}
			}

			for _, search := range entity.Searches {
				if search.Attribute == nil || search.Entity == nil {
					return xerrors.Errorf("attribute %s not found for %s search in %s entity %s namespace", search.AttrName, search.Name, entity.Name, nsName)
				}
			}

			if entity.VTEntity != nil {
				for _, attr := range entity.VTEntity.Attributes {
					if attr.AttrName == "" && attr.SearchName == "" {
						return xerrors.Errorf("attribute or search is not listed for %s vt attribute in %s entity %s namespace", attr.Name, entity.Name, nsName)
					}

					if attr.AttrName != "" {
						if entity.AttributeByName(attr.AttrName) == nil {
							return xerrors.Errorf("attribute %s not found for %s vt attribute in %s entity %s namespace", attr.AttrName, attr.Name, entity.Name, nsName)
						}
					}

					if attr.SearchName != "" {
						if entity.SearchByName(attr.SearchName) == nil && entity.AttributeByName(attr.SearchName) == nil {
							return xerrors.Errorf("search %s not found for %s vt attribute in %s entity %s namespace", attr.SearchName, attr.Name, entity.Name, nsName)
						}
					}
				}
			}
		}
	}

	return nil
}

// EntitiesMap returns map of mfd.Entities by namespace
func (p *Project) Entity(entity string) *Entity {
	for _, p := range p.Namespaces {
		if e := p.Entity(entity); e != nil {
			return e
		}
	}
	return nil
}

// Namespace returns mfd.Namespace by its name or creates if not exists
func (p *Project) Namespace(namespace string) *Namespace {
	for _, ns := range p.Namespaces {
		if ns.Name == namespace {
			return ns
		}
	}

	ns := NewNamespace(namespace)
	p.Namespaces = append(p.Namespaces, ns)
	p.NamespaceNames = append(p.NamespaceNames, namespace)

	return ns
}

func (p *Project) AddEntity(namespace string, entity *Entity) *Entity {
	ns := p.Namespace(namespace)
	return ns.AddEntity(entity)
}

func (p *Project) AddVTEntity(namespace string, entity *VTEntity) *VTEntity {
	ns := p.Namespace(namespace)
	return ns.AddVTEntity(entity)
}

func (p *Project) SuggestArrayLinks() {
	for _, namespace := range p.Namespaces {
		for _, entity := range namespace.Entities {
			// making fk links for arrays
			for _, attr := range entity.Attributes {
				if attr.ForeignKey != "" || !strings.HasSuffix(attr.Name, util.IDs) || !attr.IsArray {
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
				if attr.ForeignKey == "" {
					continue
				}

				if foreign := p.Entity(attr.ForeignKey); foreign != nil {
					attr.ForeignEntity = foreign
				}
			}

			// making search links
			for _, search := range entity.Searches {
				if attr := entity.AttributeByName(search.AttrName); attr != nil {
					search.Attribute = attr
					search.Entity = entity
				}

				if strings.Index(search.AttrName, ".") != -1 {
					parts := strings.SplitN(search.AttrName, ".", 2)
					entityName, columnName := util.EntityName(parts[0]), util.ColumnName(parts[1])
					for _, attr := range entity.Attributes {
						fkName := FKName(attr.DBName)
						if fkName == entityName && attr.ForeignEntity != nil {
							for _, a := range attr.ForeignEntity.Attributes {
								if a.Name == columnName {
									search.Attribute = a
									search.Entity = attr.ForeignEntity
								}
							}
						}
					}
				}
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

	Entities Entities `xml:"Entities>Entity"`
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
		if e.Name == entity {
			return e
		}
	}

	return nil
}

// AddNamespace add new Namespace to existing
func (n *Namespace) AddEntity(entity *Entity) *Entity {
	if existing := n.Entity(entity.Name); existing != nil {
		if existing.VTEntity == nil {
			existing.VTEntity = &VTEntity{}
		}

		if entity.VTEntity != nil {
			existing.VTEntity.Merge(entity.VTEntity)
		}

		existing.Merge(entity)
		return existing
	}

	n.Entities = append(n.Entities, entity)
	return entity
}

func (n *Namespace) UpdateFK() {

}

func (n *Namespace) VTEntities() VTEntities {
	var vtEntities VTEntities
	for _, e := range n.Entities {
		if e.VTEntity != nil {
			vtEntities = append(vtEntities, e.VTEntity)
		}
	}

	return vtEntities
}

func (n *Namespace) VTEntity(entity string) *VTEntity {
	for _, e := range n.VTEntities() {
		if e.Name == entity {
			return e
		}
	}

	return nil
}

func (n *Namespace) AddVTEntity(entity *VTEntity) *VTEntity {
	if existing := n.VTEntity(entity.Name); existing != nil {
		existing.Merge(entity)
		entity = existing
	}

	if ent := n.Entity(entity.Name); ent != nil {
		ent.VTEntity = entity
	}

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

	VTEntity *VTEntity `xml:"-"`
}

func (e *Entity) AttributeByName(name string) *Attribute {
	for _, a := range e.Attributes {
		if a.Name == name {
			return a
		}
	}

	return nil
}

// Attribute gets mfd.Attribute by its db name and type
func (e *Entity) AttributeByDBName(dbName, dbTyp string) *Attribute {
	for _, a := range e.Attributes {
		if a.DBName == dbName && a.DBType == dbTyp {
			return a
		}
	}

	return nil
}

func (e *Entity) SearchByName(name string) *Search {
	for _, s := range e.Searches {
		if s.Name == name {
			return s
		}
	}

	return nil
}

func (e *Entity) SearchByAttrName(attrName, searchType string) *Search {
	for _, s := range e.Searches {
		if s.AttrName == attrName && s.SearchType == searchType {
			return s
		}
	}

	return nil
}

// Merge fills entity from file with attributes from db
func (e *Entity) Merge(with *Entity) {
	for _, toAdd := range with.Attributes {
		if existing := e.AttributeByDBName(toAdd.DBName, toAdd.DBType); existing != nil {
			// updating exiting
			existing.Merge(toAdd)
		} else {
			// adding new
			with.Attributes = append(with.Attributes, toAdd)
		}
	}

	for _, toAdd := range with.Searches {
		// adding only new
		if existing := e.SearchByAttrName(toAdd.AttrName, toAdd.SearchType); existing == nil {
			e.Searches = append(e.Searches, toAdd)
		}
	}
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

func (e *Entity) TitleVTAttribute() *VTAttribute {
	var pkName string
	var pkAttr *VTAttribute
	if pks := e.PKs(); len(pks) > 0 {
		pkName = pks[0].Name
	}

	for _, attr := range e.VTEntity.Attributes {
		if attr.AttrName == pkName && pkAttr == nil {
			pkAttr = attr
		}
		if attr.Name == "Title" || attr.Name == "Name" || attr.Name == "Login" || attr.Name == "Alias" {
			return attr
		}
	}

	return pkAttr
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
	a.Name = with.Name
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

// convenient types
type Namespaces []*Namespace

func (ns Namespaces) Namespace(name string) *Namespace {
	for _, n := range ns {
		if n.Name == name {
			return n
		}
	}

	return nil
}

type Entities []*Entity

type Attributes []*Attribute

type Searches []*Search

func IsStatus(name string) bool {
	return strings.ToLower(name) == "statusid" || strings.ToLower(name) == "status"
}

func IsArraySearch(search string) bool {
	return search == SearchArray || search == SearchNotArray
}
