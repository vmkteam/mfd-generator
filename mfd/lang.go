package mfd

import (
	"encoding/json"
	"encoding/xml"
)

// Translations
type Translation struct {
	XMLName    xml.Name               `xml:"Translation" json:"-"`
	XMLxsi     string                 `xml:"xmlns:xsi,attr"`
	XMLxsd     string                 `xml:"xmlns:xsd,attr"`
	Language   string                 `xml:"Language"`
	Namespaces []TranslationNamespace `xml:"Namespaces>Namespace" json:"-"`
}

func (t *Translation) Namespace(namespace string) *TranslationNamespace {
	for _, ns := range t.Namespaces {
		if ns.Name == namespace {
			return &ns
		}
	}

	return nil
}

func (t *Translation) Entity(namespace, entity string) TranslationEntity {
	if ns := t.Namespace(namespace); ns != nil {
		for _, e := range ns.Entities {
			if e.Name == entity {
				return e
			}
		}
	}

	return TranslationEntity{}
}

func (t *Translation) Merge(translation Translation) {
	for _, ns := range translation.Namespaces {
		if existing := t.Namespace(ns.Name); existing != nil {
			existing.Merge(ns)
		} else {
			t.Namespaces = append(t.Namespaces, ns)
		}
	}
}

type TranslationNamespace struct {
	XMLName  xml.Name            `xml:"Namespace" json:"-"`
	Name     string              `xml:"Name,attr"`
	Entities []TranslationEntity `xml:"Entities>Entity"`
}

func (n TranslationNamespace) Entity(entity string) *TranslationEntity {
	for _, e := range n.Entities {
		if e.Name == entity {
			return &e
		}
	}

	return nil
}

func (n *TranslationNamespace) Merge(namespace TranslationNamespace) {
	for _, e := range namespace.Entities {
		if existing := n.Entity(e.Name); existing != nil {
			existing.Merge(e)
		} else {
			n.Entities = append(n.Entities, e)
		}
	}
}

type TranslationEntity struct {
	XMLName xml.Name        `xml:"Entity"`
	Name    string          `xml:"Name,attr"`
	Key     string          `xml:"Key,attr"`
	Crumbs  XMLMap          `xml:"Crumbs"`
	Form    XMLMap          `xml:"Form"`
	List    TranslationList `xml:"List"`
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

func (e *TranslationEntity) Merge(entity TranslationEntity) {
	if e.List.Title == "" {
		e.List.Title = entity.List.Title
	}

	e.Crumbs = mergeMap(e.Crumbs, entity.Crumbs)
	e.Form = mergeMap(e.Form, entity.Form)
	e.List.Filter = mergeMap(e.List.Filter, entity.List.Filter)
	e.List.Headers = mergeMap(e.List.Headers, entity.List.Headers)
}

type TranslationList struct {
	Title   string `xml:"Title" json:"title"`
	Filter  XMLMap `xml:"Filter" json:"filter"`
	Headers XMLMap `xml:"Headers" json:"headers"`
}

func mergeMap(base, new map[string]string) map[string]string {
	if base == nil {
		return new
	}

	for k, v := range new {
		if e, ok := base[k]; !ok || e == "" {
			base[k] = v
		}
	}

	return base
}
