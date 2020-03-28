package vt

import "github.com/vmkteam/mfd-generator/mfd"

// PackServerNamespaces packs namespaces for zenprc server code
func PackServerNamespaces(namespaces []*mfd.VTNamespace, options Options) (NamespaceData, error) {
	var models []EntityData
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			if entity.Mode == mfd.ModeNone {
				continue
			}

			// creating entity for template
			mdl, err := PackEntity(*entity, options)
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
