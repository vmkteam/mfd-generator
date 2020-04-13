package xmllang

import (
	"github.com/vmkteam/mfd-generator/mfd"
)

func Translate(p *mfd.Project, language string) *mfd.Translation {
	tr := &mfd.Translation{
		Language:   language,
		Namespaces: []*mfd.TranslationNamespace{},
	}

	for _, ns := range p.VTNamespaces {
		tn := &mfd.TranslationNamespace{
			Name:     ns.Name,
			Entities: []*mfd.TranslationEntity{},
		}

		for _, e := range ns.Entities {
			key := mfd.VarName(e.Name)

			te := &mfd.TranslationEntity{
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

			for _, a := range e.TmplAttributes {
				key := mfd.VarName(a.Name)

				trs := mfd.Translate(language, key)

				if a.Form != "" && a.Form != mfd.TypeHTMLNone {
					te.Form.Append(key+"Label", trs)
				}

				if a.Search != "" && a.Search != mfd.TypeHTMLNone {
					te.List.Filter.Append(key, trs)
				}

				if a.List {
					// override statusId key because headers for summary, and summary have Status object
					if mfd.IsStatus(key) {
						key = "status"
					}
					te.List.Headers.Append(key, trs)
				}
			}
			te.List.Headers.Append("actions", mfd.Translate(language, "actions"))

			tn.Entities = append(tn.Entities, te)
		}

		tr.Namespaces = append(tr.Namespaces, tn)
	}

	return tr
}
