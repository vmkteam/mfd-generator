package xmllang

import (
	"github.com/vmkteam/mfd-generator/mfd"
)

func Translate(ns *mfd.VTNamespace, translation mfd.Translation, entities []string, language string) mfd.Translation {
	namespace := translation.Namespace(ns.Name)
	if namespace == nil {
		namespace = &mfd.TranslationNamespace{
			Name:     ns.Name,
			Entities: []*mfd.TranslationEntity{},
		}
	}

	for _, entityName := range entities {
		e := ns.VTEntity(entityName)
		if e == nil {
			continue
		}

		entity := namespace.Entity(e.Name)

		// deleting unused translations
		if e.Mode == mfd.ModeReadOnly || e.Mode == mfd.ModeNone {
			if entity != nil {
				namespace.DeleteEntity(e.Name)
			}

			continue
		}

		if entity == nil {
			key := mfd.VarName(e.Name)
			entity = &mfd.TranslationEntity{
				Name: e.Name,
				Key:  key,
				Form: mfd.NewXMLMap(nil),
				List: &mfd.TranslationList{
					Title:   mfd.Translate(language, mfd.MakePlural(key)),
					Filter:  mfd.NewXMLMap(map[string]string{"quickFilterPlaceholder": ""}),
					Headers: mfd.NewXMLMap(nil),
				},
				Crumbs: mfd.NewXMLMap(map[string]string{
					key + "List": mfd.Translate(language, mfd.MakePlural(key)),
					key + "Add":  mfd.Translate(language, "add"),
					key + "Edit": mfd.Translate(language, "edit"),
				}),
			}
		}

		// fix empty fields
		fixup(entity)

		for _, a := range e.TmplAttributes {
			key := mfd.VarName(a.Name)

			trs := mfd.Translate(language, key)

			// deleting unused translations
			if emptyOrNone(a.Form) || e.Mode == mfd.ModeReadOnlyWithTemplates {
				entity.Form.Delete(key + "Label")
			}

			if !emptyOrNone(a.Form) && e.Mode == mfd.ModeFull {
				entity.Form.Append(key+"Label", trs)
			}

			if emptyOrNone(a.Search) && entity.List.Filter != nil {
				entity.List.Filter.Delete(key)
			}

			if !emptyOrNone(a.Search) {
				entity.List.Filter.Append(key, trs)
			}

			if a.List {
				// override statusId key because headers for summary, and summary have Status object
				if mfd.IsStatus(key) {
					key = "status"
				}

				entity.List.Headers.Append(key, trs)
			} else {
				// override statusId key because headers for summary, and summary have Status object
				if mfd.IsStatus(key) {
					key = "status"
				}

				entity.List.Headers.Delete(key)
			}
		}

		switch e.Mode {
		case mfd.ModeFull:
			entity.List.Headers.Append("actions", mfd.Translate(language, "actions"))
		case mfd.ModeReadOnlyWithTemplates:
			entity.List.Headers.Delete("actions")

			entity.Crumbs.Delete(entity.Key + "Add")
			entity.Crumbs.Delete(entity.Key + "Edit")
		}

		namespace.AddEntity(entity)
	}

	if len(namespace.Entities) > 0 {
		translation.AddNamespace(namespace)
	}

	return translation
}

func emptyOrNone(val string) bool {
	return val == "" || val == mfd.TypeHTMLNone
}

func fixup(entity *mfd.TranslationEntity) {
	// safe fields
	if entity.Form == nil {
		entity.Form = mfd.NewXMLMap(nil)
	}
	if entity.List == nil {
		entity.List = &mfd.TranslationList{}
	}
	if entity.List.Filter == nil {
		entity.List.Filter = mfd.NewXMLMap(nil)
	}
	if entity.List.Headers == nil {
		entity.List.Headers = mfd.NewXMLMap(nil)
	}
	if entity.Crumbs == nil {
		entity.Crumbs = mfd.NewXMLMap(nil)
	}
}
