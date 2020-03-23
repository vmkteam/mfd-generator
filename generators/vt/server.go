package vt

import "github.com/vmkteam/mfd-generator/mfd"

// PackServerNamespaces packs namespaces for zenprc server code
func PackServerNamespaces(namespaces []*mfd.VTNamespace) (NamespaceData, error) {
	var models []EntityData
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			// creating entity for template
			mdl, err := PackEntity(*entity)
			if err != nil {
				return NamespaceData{}, err
			}

			models = append(models, mdl)
		}
	}

	return NamespaceData{
		Entities: models,
	}, nil
}
