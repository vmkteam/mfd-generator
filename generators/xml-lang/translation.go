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
			entity = mfd.NewTranslationEntity(e.Name, language)
		}

		// fix empty fields
		fixup(entity)

		entity.FillByVTEntity(e, language)

		namespace.AddEntity(entity)
	}

	if len(namespace.Entities) > 0 {
		translation.AddNamespace(namespace)
	}

	return translation
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
