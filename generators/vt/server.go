package vt

import "github.com/vmkteam/mfd-generator/mfd"

func NewServerPackage(namespaces mfd.Namespaces) (TemplatePackage, error) {
	var models []TemplateEntity
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			// creating entity for template
			mdl, err := NewTemplateEntity(*entity)
			if err != nil {
				return TemplatePackage{}, err
			}

			models = append(models, mdl)
		}
	}

	return TemplatePackage{
		Entities: models,
	}, nil
}
