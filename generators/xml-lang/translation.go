package xmllang

import (
	"github.com/vmkteam/mfd-generator/mfd"
)

func Translate(p *mfd.Project, translation *mfd.Translation, language string) {
	for _, ns := range p.VTNamespaces {
		namespace := translation.Namespace(ns.Name)
		if namespace == nil {
			namespace = &mfd.TranslationNamespace{
				Name:     ns.Name,
				Entities: []*mfd.TranslationEntity{},
			}
		}

		for _, e := range ns.Entities {
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
					List: mfd.TranslationList{
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

			// safe fields
			if entity.Form == nil {
				entity.Form = mfd.NewXMLMap(nil)
			}

			// deleting unused translations
			if (e.Mode == mfd.ModeReadOnly || e.Mode == mfd.ModeNone) && entity.Crumbs != nil {
				entity.Crumbs.Delete(entity.Key + "Add")
				entity.Crumbs.Delete(entity.Key + "Edit")
			}

			for _, a := range e.TmplAttributes {
				key := mfd.VarName(a.Name)

				trs := mfd.Translate(language, key)

				// deleting unused translations
				if a.Form == "" || e.Mode == mfd.ModeReadOnlyWithTemplates {
					entity.Form.Delete(key + "Label")
				}

				if a.Form != "" && e.Mode == mfd.ModeFull {
					entity.Form.Append(key+"Label", trs)
				}

				if a.Search != "" {
					if entity.List.Filter == nil {
						entity.List.Filter = mfd.NewXMLMap(nil)
					}
					entity.List.Filter.Append(key, trs)
				}

				if a.List {
					// override statusId key because headers for summary, and summary have Status object
					if mfd.IsStatus(key) {
						key = "status"
					}
					if entity.List.Headers == nil {
						entity.List.Headers = mfd.NewXMLMap(nil)
					}
					entity.List.Headers.Append(key, trs)
				}
			}

			switch e.Mode {
			case mfd.ModeFull:
				entity.List.Headers.Append("actions", mfd.Translate(language, "actions"))
			case mfd.ModeReadOnlyWithTemplates:
				entity.List.Headers.Delete("actions")
			}

			namespace.AddEntity(entity)
		}

		translation.AddNamespace(namespace)
	}
}
