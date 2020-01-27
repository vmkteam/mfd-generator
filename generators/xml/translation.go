package xml

import "github.com/vmkteam/mfd-generator/mfd"

func Translate(p *mfd.Project, language string) mfd.Translation {
	tr := mfd.Translation{
		Language:   language,
		Namespaces: []mfd.TranslationNamespace{},
	}

	for _, ns := range p.Namespaces {
		tn := mfd.TranslationNamespace{
			Name:     ns.Name,
			Entities: []mfd.TranslationEntity{},
		}

		for _, e := range ns.Entities {
			key := mfd.VarName(e.Name)

			te := mfd.TranslationEntity{
				Name: e.Name,
				Key:  key,
				Form: mfd.XMLMap{},
				List: mfd.TranslationList{
					Title:   mfd.Translate(language, mfd.MakePlural(key)),
					Filter:  mfd.XMLMap{"quickFilterPlaceholder": ""},
					Headers: mfd.XMLMap{},
				},
				Crumbs: map[string]string{
					key + "List": mfd.Translate(language, mfd.MakePlural(key)),
					key + "Add":  mfd.Translate(language, "add"),
					key + "Edit": mfd.Translate(language, "edit"),
				},
			}

			for _, a := range e.VTEntity.TmplAttributes {
				key := mfd.VarName(a.Name)

				trs := mfd.Translate(language, key)

				if a.Form != "" && a.Form != mfd.TypeHTMLNone {
					te.Form[key+"Label"] = trs
				}

				if a.Search != "" && a.Search != mfd.TypeHTMLNone {
					te.List.Filter[key] = trs
				}

				if a.List {
					// override statusId key because headers for summary, and summary have Status object
					if mfd.IsStatus(key) {
						key = "status"
					}
					te.List.Headers[key] = trs
				}
			}
			te.List.Headers["actions"] = mfd.Translate(language, "actions")

			tn.Entities = append(tn.Entities, te)
		}

		tr.Namespaces = append(tr.Namespaces, tn)
	}

	return tr
}
