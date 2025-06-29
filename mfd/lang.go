package mfd

import (
	"encoding/xml"
)

type Translation struct {
	XMLName    xml.Name                `xml:"Translation" json:"-"`
	XMLxsi     string                  `xml:"xmlns:xsi,attr" json:"-"`
	XMLxsd     string                  `xml:"xmlns:xsd,attr" json:"-"`
	Language   string                  `xml:"Language" json:"language"`
	Namespaces []*TranslationNamespace `xml:"Namespaces>Namespace" json:"namespaces"`
}

func (t *Translation) Namespace(namespace string) *TranslationNamespace {
	for _, ns := range t.Namespaces {
		if ns.Name == namespace {
			return ns
		}
	}

	return nil
}

func (t *Translation) Entity(namespace, entity string) *TranslationEntity {
	if ns := t.Namespace(namespace); ns != nil {
		for _, e := range ns.Entities {
			if e.Name == entity {
				return e
			}
		}
	}

	return nil
}

func (t *Translation) AddNamespace(namespace *TranslationNamespace) {
	for i, n := range t.Namespaces {
		if n.Name == namespace.Name {
			t.Namespaces[i] = namespace
			return
		}
	}

	t.Namespaces = append(t.Namespaces, namespace)
}

type TranslationNamespace struct {
	XMLName  xml.Name             `xml:"Namespace" json:"-"`
	Name     string               `xml:"Name,attr" json:"name"`
	Entities []*TranslationEntity `xml:"Entities>Entity"  json:"entities"`
}

func (n TranslationNamespace) Entity(entity string) *TranslationEntity {
	for _, e := range n.Entities {
		if e.Name == entity {
			return e
		}
	}

	return nil
}

func (n *TranslationNamespace) AddEntity(entity *TranslationEntity) {
	for i, e := range n.Entities {
		if e.Name == entity.Name {
			n.Entities[i] = entity
			return
		}
	}

	n.Entities = append(n.Entities, entity)
}

func (n *TranslationNamespace) DeleteEntity(entity string) {
	for i, e := range n.Entities {
		if e.Name == entity {
			n.Entities = append(n.Entities[:i], n.Entities[i+1:]...)
			return
		}
	}
}

type TranslationEntity struct {
	XMLName xml.Name         `xml:"Entity" json:"-"`
	Name    string           `xml:"Name,attr" json:"name"`
	Key     string           `xml:"Key,attr" json:"key"`
	Crumbs  *XMLMap          `xml:"Crumbs" json:"crumbs"`
	Form    *XMLMap          `xml:"Form" json:"form"`
	List    *TranslationList `xml:"List" json:"list"`
}

func NewTranslationEntity(entityName, language string) *TranslationEntity {
	key := VarName(entityName)
	return &TranslationEntity{
		Name: entityName,
		Key:  key,
		Form: NewXMLMap(nil),
		List: &TranslationList{
			Title:   Translate(language, MakePlural(key)),
			Filter:  NewXMLMap(map[string]string{"quickFilterPlaceholder": ""}),
			Headers: NewXMLMap(nil),
		},
		Crumbs: NewXMLMap(map[string]string{
			key + "List": Translate(language, MakePlural(key)),
			key + "Add":  Translate(language, "add"),
			key + "Edit": Translate(language, "edit"),
		}),
	}
}

func (e *TranslationEntity) ToJSONMap() map[string]interface{} {
	jsM := map[string]interface{}{
		"breadcrumbs": e.Crumbs,
		e.Key: map[string]interface{}{
			"form": e.Form,
			"list": e.List,
		},
	}

	return jsM
}

func (e *TranslationEntity) FillByVTEntity(vte *VTEntity, language string) {
	for _, a := range vte.TmplAttributes {
		key := VarName(a.Name)

		trs := Translate(language, key)

		// deleting unused translations
		if emptyOrNone(a.Form) || vte.Mode == ModeReadOnlyWithTemplates {
			e.Form.Delete(key + "Label")
		}

		if !emptyOrNone(a.Form) && vte.Mode == ModeFull {
			e.Form.Append(key+"Label", trs)
		}

		if emptyOrNone(a.Search) && e.List.Filter != nil {
			e.List.Filter.Delete(key)
		}

		if !emptyOrNone(a.Search) {
			e.List.Filter.Append(key, trs)
		}

		if a.List {
			// override statusId key because headers for summary, and summary have Status object
			if IsStatus(key) {
				key = "status"
			}

			e.List.Headers.Append(key, trs)
		} else {
			// override statusId key because headers for summary, and summary have Status object
			if IsStatus(key) {
				key = "status"
			}

			e.List.Headers.Delete(key)
		}
	}

	switch vte.Mode {
	case ModeFull:
		e.List.Headers.Append("actions", Translate(language, "actions"))
	case ModeReadOnlyWithTemplates:
		e.List.Headers.Delete("actions")

		e.Crumbs.Delete(e.Key + "Add")
		e.Crumbs.Delete(e.Key + "Edit")
	}
}

type TranslationList struct {
	Title   string  `xml:"Title" json:"title"`
	Filter  *XMLMap `xml:"Filter" json:"filter"`
	Headers *XMLMap `xml:"Headers" json:"headers"`
}

func emptyOrNone(val string) bool {
	return val == "" || val == TypeHTMLNone
}
