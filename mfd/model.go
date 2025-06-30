package mfd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

var Version = "unknown"

// go-pg versions
const (
	GoPG8  = 8
	GoPG9  = 9
	GoPG10 = 10
)

// nullable options
const (
	NullableYes   = "Yes"
	NullableNo    = "No"
	NullableEmpty = "CheckEmpty"
)

// vfsFile entity name
const (
	VfsFile      = "VfsFile"
	JSONFieldSep = "->"
)

type SearchType string

const (
	SearchEquals               SearchType = "SEARCHTYPE_EQUALS"
	SearchNotEquals            SearchType = "SEARCHTYPE_NOT_EQUALS"
	SearchNull                 SearchType = "SEARCHTYPE_NULL"
	SearchNotNull              SearchType = "SEARCHTYPE_NOT_NULL"
	SearchGE                   SearchType = "SEARCHTYPE_GE"
	SearchLE                   SearchType = "SEARCHTYPE_LE"
	SearchG                    SearchType = "SEARCHTYPE_G"
	SearchL                    SearchType = "SEARCHTYPE_L"
	SearchLeftLike             SearchType = "SEARCHTYPE_LEFT_LIKE"
	SearchLeftILike            SearchType = "SEARCHTYPE_LEFT_ILIKE"
	SearchRightLike            SearchType = "SEARCHTYPE_RIGHT_LIKE"
	SearchRightILike           SearchType = "SEARCHTYPE_RIGHT_ILIKE"
	SearchLike                 SearchType = "SEARCHTYPE_LIKE"
	SearchILike                SearchType = "SEARCHTYPE_ILIKE"
	SearchArray                SearchType = "SEARCHTYPE_ARRAY"
	SearchNotArray             SearchType = "SEARCHTYPE_NOT_INARRAY"
	SearchTypeArrayContains    SearchType = "SEARCHTYPE_ARRAY_CONTAINS"
	SearchTypeArrayNotContains SearchType = "SEARCHTYPE_ARRAY_NOT_CONTAINS"
	SearchTypeArrayContained   SearchType = "SEARCHTYPE_ARRAY_CONTAINED"
	SearchTypeArrayIntersect   SearchType = "SEARCHTYPE_ARRAY_INTERSECT"
	SearchTypeJsonbPath        SearchType = "SEARCHTYPE_JSONB_PATH"
)

func (si SearchType) String() string {
	return string(si)
}

func (si SearchType) IsArraySearch() bool {
	return si == SearchArray || si == SearchNotArray
}

type FilterType struct {
	Name    string
	Exclude bool
	IsArray bool
}

func (ft FilterType) ExcludeString() string {
	return boolAsString(ft.Exclude)
}

var (
	FilterTypeBySearchType = map[SearchType]FilterType{
		SearchEquals:               {Name: "SearchTypeEquals"},
		SearchNotEquals:            {Name: "SearchTypeEquals", Exclude: true},
		SearchNull:                 {Name: "SearchTypeNull"},
		SearchNotNull:              {Name: "SearchTypeNull", Exclude: true},
		SearchGE:                   {Name: "SearchTypeGE"},
		SearchLE:                   {Name: "SearchTypeLE"},
		SearchG:                    {Name: "SearchTypeGreater"},
		SearchL:                    {Name: "SearchTypeLess"},
		SearchLeftLike:             {Name: "SearchTypeLLike"},
		SearchLeftILike:            {Name: "SearchTypeLILike"},
		SearchRightLike:            {Name: "SearchTypeRLike"},
		SearchRightILike:           {Name: "SearchTypeRILike"},
		SearchLike:                 {Name: "SearchTypeLike"},
		SearchILike:                {Name: "SearchTypeILike"},
		SearchArray:                {Name: "SearchTypeArray", IsArray: true},
		SearchNotArray:             {Name: "SearchTypeArray", Exclude: true, IsArray: true},
		SearchTypeArrayContains:    {Name: "SearchTypeArrayContains"},
		SearchTypeArrayNotContains: {Name: "SearchTypeArrayContains", Exclude: true},
		SearchTypeArrayContained:   {Name: "SearchTypeArrayContained", IsArray: true},
		SearchTypeArrayIntersect:   {Name: "SearchTypeArrayIntersect", IsArray: true},
		SearchTypeJsonbPath:        {Name: "SearchTypeJsonbPath"},
	}
)

type CustomType struct {
	DBType   string `xml:"DBType,attr,omitempty" json:"dbType"`
	GoType   string `xml:"GoType,attr,omitempty" json:"goType"`
	GoImport string `xml:"GoImport,attr,omitempty" json:"goImport"`
}

type CustomTypes []CustomType

func (c CustomTypes) GoImport(goType, dbType string) (string, bool) {
	for _, customType := range c {
		if Element(customType.GoType) == Element(goType) && (customType.DBType == dbType || customType.DBType == "*") {
			return customType.GoImport, true
		}
	}

	return "", false
}

type NSMapping struct {
	Namespace string `json:"namespace"`
	Entity    string `json:"entity"`
}

// Project is xml element
type Project struct {
	XMLName        xml.Name    `xml:"Project" json:"-"`
	XMLxsi         string      `xml:"xmlns:xsi,attr" json:"-"`
	XMLxsd         string      `xml:"xmlns:xsd,attr" json:"-"`
	Name           string      `json:"name"`
	NamespaceNames []string    `xml:"PackageNames>string" json:"-"`
	Languages      []string    `xml:"Languages>string" json:"languages"`
	GoPGVer        int         `xml:"GoPGVer" json:"goPGVer"`
	CustomTypes    CustomTypes `xml:"CustomTypes>CustomType,omitempty" json:"customTypes,omitempty"`
	Dictionary     *Dictionary `xml:"Dictionary" json:"dict,omitempty"`

	Namespaces   []*Namespace   `xml:"-" json:"-"`
	VTNamespaces []*VTNamespace `xml:"-" json:"-"`
	NSMapping    []NSMapping    `xml:"-" json:"namespaces"`
}

type Dictionary struct {
	Entries []Entry `xml:",any"`
}

type Entry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func NewProject(name string, goPGVer int) *Project {
	return &Project{
		Name:           name,
		NamespaceNames: []string{},

		GoPGVer:   goPGVer,
		Languages: []string{EnLang},

		XMLxsi: "",
		XMLxsd: "",
	}
}

// Namespace returns mfd.Namespace by its name
func (p *Project) Namespace(namespace string) *Namespace {
	for _, ns := range p.Namespaces {
		if strings.EqualFold(ns.Name, namespace) {
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

// EntityByTable returns mfd.Entity by its table
func (p *Project) EntityByTable(table string) *Entity {
	for _, n := range p.Namespaces {
		if e := n.EntityByTable(table); e != nil {
			return e
		}
	}
	return nil
}

// VTEntity returns mfd.VTEntity by its name
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
	if p.GoPGVer < GoPG8 || p.GoPGVer > GoPG10 {
		return fmt.Errorf("unsupported go-pg version: %d", p.GoPGVer)
	}

	for _, nsName := range p.NamespaceNames {
		ns := p.Namespace(nsName)
		if ns == nil {
			return fmt.Errorf("namespace %s listed in names but not found", nsName)
		}

		for _, entity := range ns.Entities {
			if err := p.IsConsistentEntity(entity, nsName); err != nil {
				return err
			}
		}
	}

	for _, vtNamespace := range p.VTNamespaces {
		ns := p.Namespace(vtNamespace.Name)
		if ns == nil {
			return fmt.Errorf("namespace %s not found for vt", vtNamespace.Name)
		}

		for _, vtEntity := range vtNamespace.Entities {
			if err := p.IsConsistentVTEntity(vtEntity, vtNamespace.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Project) IsConsistentEntity(entity *Entity, namespace string) error {
	for _, attr := range entity.Attributes {
		if attr.ForeignKey != "" && attr.ForeignEntity == nil {
			return fmt.Errorf("fk entity %s not found for %s column in %s entity %s namespace", attr.ForeignKey, attr.Name, entity.Name, namespace)
		}
	}

	for _, search := range entity.Searches {
		if search.Attribute == nil || search.Entity == nil {
			return fmt.Errorf("attribute %s not found for %s search in %s entity %s namespace", search.AttrName, search.Name, entity.Name, namespace)
		}
	}
	return nil
}

func (p *Project) IsConsistentVTEntity(vtEntity *VTEntity, namespace string) error {
	if vtEntity.Entity == nil {
		return fmt.Errorf("entity not found vtEntity %s in %s namespace", vtEntity.Name, namespace)
	}

	for _, vtAttribute := range vtEntity.Attributes {
		if vtAttribute.Attribute == nil {
			if vtAttribute.AttrName != "" {
				return fmt.Errorf("attribute %s not found for attribute %s in vtEntity %s in %s namespace", vtAttribute.AttrName, vtAttribute.Name, vtEntity.Name, namespace)
			}
			if vtAttribute.SearchName != "" {
				return fmt.Errorf("search %s not found for attribute %s in vtEntity %s in %s namespace", vtAttribute.SearchName, vtAttribute.Name, vtEntity.Name, namespace)
			}
		}
	}
	return nil
}

func (p *Project) ValidateNames() error {
	var errors []string

	for _, namespace := range p.Namespaces {
		if IsReserved(namespace.Name) || IsReservedByMFD(namespace.Name) {
			errors = append(errors, fmt.Sprintf(`namspace name: "%s" is reserved`, namespace.Name))
		}

		for _, entity := range namespace.Entities {
			if IsReserved(entity.Name) || IsReservedByMFD(entity.Name) {
				errors = append(errors, fmt.Sprintf(`entity name: "%s" is reserved`, entity.Name))
			}
		}
	}

	for _, namespace := range p.VTNamespaces {
		if IsReserved(namespace.Name) || IsReservedByMFD(namespace.Name) {
			errors = append(errors, fmt.Sprintf(`namspace name: "%s" is reserved`, namespace.Name))
		}

		for _, entity := range namespace.Entities {
			// skip if read only or none
			if entity.Mode == ModeReadOnly || entity.Mode == ModeNone {
				continue
			}

			if IsReserved(entity.Name) || IsReservedByMFD(entity.Name) {
				errors = append(errors, fmt.Sprintf(`vt entity name: "%s" is reserved`, entity.Name))
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return fmt.Errorf("invalid names detected (%d):\n%s", len(errors), strings.Join(errors, "\n"))
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
			p.updateForeignAttr(entity)
			// making search links
			p.updateSearchLinks(entity)
		}
	}

	// making vt links
	for _, vtNamespace := range p.VTNamespaces {
		for _, vtEntity := range vtNamespace.Entities {
			// Backward compatibility
			if vtEntity.NoTemplates && vtEntity.Mode == "" {
				vtEntity.NoTemplates = false
				vtEntity.Mode = ModeReadOnly
			}

			if vtEntity.Mode == "" {
				vtEntity.Mode = ModeFull
			}

			if entity := p.Entity(vtEntity.Name); entity != nil {
				vtEntity.Entity = entity
				vtEntity.Attributes.updateAttr(entity)
			}

			for _, tmpl := range vtEntity.TmplAttributes {
				tmpl.VTAttribute = vtEntity.Attribute(tmpl.AttrName)
			}
		}
	}
}

func (p *Project) updateForeignAttr(entity *Entity) {
	for _, attr := range entity.Attributes {
		if foreign := p.Entity(attr.ForeignKey); foreign != nil {
			attr.ForeignEntity = foreign
		}
	}
}

func (p *Project) updateSearchLinks(entity *Entity) {
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

func (p *Project) AddCustomTypes(mapping model.CustomTypeMapping) (newCustomTypes CustomTypes) {
	for _, customType := range mapping {
		if customType.GoType == "" {
			continue
		}

		existed := false
		for i, existing := range p.CustomTypes {
			if existing.DBType != "" && existing.DBType == customType.PGType {
				p.CustomTypes[i] = CustomType{
					DBType:   customType.PGType,
					GoType:   customType.GoType,
					GoImport: customType.GoImport,
				}

				existed = true
				break
			}
		}

		if !existed {
			ct := CustomType{
				DBType:   customType.PGType,
				GoType:   customType.GoType,
				GoImport: customType.GoImport,
			}

			p.CustomTypes = append(p.CustomTypes, ct)
			newCustomTypes = append(newCustomTypes, ct)
		}
	}

	return newCustomTypes
}

func (p *Project) CustomTypeMapping() model.CustomTypeMapping {
	ctm := model.CustomTypeMapping{}
	for _, customType := range p.CustomTypes {
		if customType.DBType != "" {
			ctm.Add(customType.DBType, customType.GoType, customType.GoImport)
		}
	}

	return ctm
}

func (p *Project) NamespacesMapping() []NSMapping {
	var result []NSMapping

	for _, ns := range p.Namespaces {
		for _, e := range ns.Entities {
			result = append(result, NSMapping{
				Namespace: ns.Name,
				Entity:    e.Name,
			})
		}
	}

	return result
}

func (p *Project) UpdateByNSMapping() {
	for _, mapping := range p.NSMapping {
		entity := p.Entity(mapping.Entity)
		if entity == nil {
			continue
		}

		source := p.Namespace(entity.Namespace)
		target := p.Namespace(mapping.Namespace)
		if source == nil || target == nil {
			continue
		}

		if source.Name == target.Name {
			continue
		}

		entity.Namespace = target.Name
		target.Entities = append(target.Entities, entity)
		index := source.EntityIndex(entity.Name)
		source.Entities = append(source.Entities[:index], source.Entities[index+1:]...)
	}
}

func (p *Project) MarshalJSON() ([]byte, error) {
	type aux Project
	a := aux(*p)
	a.NSMapping = p.NamespacesMapping()

	return json.Marshal(a)
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

		XMLxsi: "",
		XMLxsd: "",
	}
}

// Entity returns mfd.Entity by its name
func (n *Namespace) Entity(entity string) *Entity {
	for _, e := range n.Entities {
		if strings.EqualFold(e.Name, entity) {
			return e
		}
	}

	return nil
}

// EntityByTable returns mfd.Entity by table name
func (n *Namespace) EntityByTable(table string) *Entity {
	for _, e := range n.Entities {
		if e.Table == table {
			return e
		}
	}

	return nil
}

// EntityIndex returns mfd.Entity index by its name
func (n *Namespace) EntityIndex(entity string) int {
	for i, e := range n.Entities {
		if strings.EqualFold(e.Name, entity) {
			return i
		}
	}

	return -1
}

// EntityNames returns all entity names
func (n *Namespace) EntityNames() []string {
	result := make([]string, len(n.Entities))
	for i, e := range n.Entities {
		result[i] = e.Name
	}

	return result
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
	Name      string `xml:"Name,attr" json:"name"`
	Namespace string `xml:"Namespace,attr" json:"namespace"`
	Table     string `xml:"Table,attr" json:"table"`

	Attributes Attributes `xml:"Attributes>Attribute,omitempty" json:"attributes"`
	Searches   Searches   `xml:"Searches>Search,omitempty" json:"searches"`
}

// AttributeByName gets mfd.Attribute by its name
func (e *Entity) AttributeByName(name string) *Attribute {
	if IsJSON(name) {
		s := strings.Split(name, JSONFieldSep)
		name = s[0]
	}
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
func (e *Entity) SearchByAttrName(attrName string, searchType SearchType) *Search {
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
	Name   string `xml:"Name,attr" json:"name"`
	DBName string `xml:"DBName,attr" json:"dbName"`

	IsArray        bool   `xml:"IsArray,attr,omitempty" json:"isArray"`
	DisablePointer bool   `xml:"DisablePointer,attr,omitempty" json:"disablePointer"`
	DBType         string `xml:"DBType,attr,omitempty" json:"dbType"`
	GoType         string `xml:"GoType,attr,omitempty" json:"goType"`

	PrimaryKey    bool    `xml:"PK,attr" json:"pk"`
	ForeignKey    string  `xml:"FK,attr,omitempty" json:"fk"`
	ForeignEntity *Entity `xml:"-" json:"-"`

	Null      string `xml:"Nullable,attr" json:"nullable"`
	Addable   *bool  `xml:"Addable,attr" json:"addable"`
	Updatable *bool  `xml:"Updatable,attr" json:"updatable"`
	Min       int    `xml:"Min,attr" json:"min"`
	Max       int    `xml:"Max,attr" json:"max"`
	Default   string `xml:"Default,attr,omitempty" json:"defaultVal"`
}

// Merge fills attribute (from file) values from db
func (a *Attribute) Merge(with *Attribute, overwriteGoType bool) {
	// a.Name = with.Name
	a.DBName = with.DBName
	a.DBType = with.DBType
	a.IsArray = with.IsArray
	a.ForeignKey = with.ForeignKey
	a.ForeignEntity = with.ForeignEntity
	a.Max = with.Max
	a.Min = with.Min

	if a.Addable == nil {
		addable := true
		a.Addable = &addable
	}

	if a.Updatable == nil {
		updatable := true
		a.Updatable = &updatable
	}

	if overwriteGoType {
		a.GoType = with.GoType
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

// Search is xml element
type Search struct {
	XMLName    xml.Name   `xml:"Search" json:"-"`
	Name       string     `xml:"Name,attr" json:"name"`
	AttrName   string     `xml:"AttrName,attr" json:"attrName"`
	SearchType SearchType `xml:"SearchType,attr" json:"searchType"`
	GoType     string     `xml:"GoType,attr,omitempty" json:"goType"`

	Attribute *Attribute `xml:"-" json:"-"`
	Entity    *Entity    `xml:"-" json:"-"`
}

func (s *Search) IsForeignSearch() bool {
	return strings.Contains(s.AttrName, ".")
}

func (s *Search) ForeignAttribute() (entity, attribute string) {
	parts := strings.SplitN(s.AttrName, ".", 2)
	return util.EntityName(parts[0]), util.ColumnName(parts[1])
}

func IsStatus(name string) bool {
	return strings.EqualFold(name, "statusid") || strings.EqualFold(name, "status_id")
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
func (a Attributes) Merge(attr *Attribute, overwriteGoType bool) (Attributes, *Attribute) {
	for i, existing := range a {
		if existing.DBName == attr.DBName && existing.DBType == attr.DBType {
			existing.Merge(attr, overwriteGoType)
			a[i] = existing
			return a, existing
		}
	}

	return append(a, attr), attr
}

const (
	trueS  = "true"
	falseS = "false"
)

func boolAsString(b bool) string {
	if b {
		return trueS
	}

	return falseS
}
