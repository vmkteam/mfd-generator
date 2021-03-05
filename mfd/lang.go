package mfd

import (
	"encoding/json"
	"encoding/xml"
)

// Translations
type Translation struct {
	XMLName    xml.Name                `xml:"Translation" json:"-"`
	XMLxsi     string                  `xml:"xmlns:xsi,attr"`
	XMLxsd     string                  `xml:"xmlns:xsd,attr"`
	Language   string                  `xml:"Language"`
	Namespaces []*TranslationNamespace `xml:"Namespaces>Namespace" json:"-"`
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
	Name     string               `xml:"Name,attr"`
	Entities []*TranslationEntity `xml:"Entities>Entity"`
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
	XMLName xml.Name         `xml:"Entity"`
	Name    string           `xml:"Name,attr"`
	Key     string           `xml:"Key,attr"`
	Crumbs  *XMLMap          `xml:"Crumbs"`
	Form    *XMLMap          `xml:"Form"`
	List    *TranslationList `xml:"List"`
}

func (e TranslationEntity) MarshalJSON() ([]byte, error) {
	jsM := map[string]interface{}{
		"breadcrumbs": e.Crumbs,
		e.Key: map[string]interface{}{
			"form": e.Form,
			"list": e.List,
		},
	}

	return json.Marshal(jsM)
}

type TranslationList struct {
	Title   string  `xml:"Title,omitempty" json:"title"`
	Filter  *XMLMap `xml:"Filter" json:"filter"`
	Headers *XMLMap `xml:"Headers" json:"headers"`
}
