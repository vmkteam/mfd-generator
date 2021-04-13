package api

import (
	"github.com/vmkteam/mfd-generator/mfd"
)

func newProject(project *mfd.Project) *Project {
	if project == nil {
		return nil
	}

	return &Project{
		Name:        project.Name,
		Languages:   project.Languages,
		GoPGVer:     project.GoPGVer,
		CustomTypes: newCustomTypes(project.CustomTypes),
		Namespaces:  newNamespaces(project.Namespaces),
	}
}

func newNamespaces(namespaces []*mfd.Namespace) []Namespace {
	var x []Namespace
	for i := range namespaces {
		if namespaces[i] == nil {
			continue
		}
		x = append(x, newNamespace(*namespaces[i]))
	}
	return x
}

func newNamespace(namespace mfd.Namespace) Namespace {
	return Namespace{
		Name:     namespace.Name,
		Entities: newEntities(namespace.Entities),
	}
}

func newEntities(entities []*mfd.Entity) []Entity {
	var x []Entity
	for i := range entities {
		if entities[i] == nil {
			continue
		}
		x = append(x, newEntity(*entities[i]))
	}
	return x
}

func newEntity(entity mfd.Entity) Entity {
	return Entity{
		Name:       entity.Name,
		Namespace:  entity.Namespace,
		Table:      entity.Table,
		Attributes: newAttributes(entity.Attributes),
		Searches:   newSearches(entity.Searches),
	}
}

func newAttributes(attributes mfd.Attributes) []Attribute {
	var x []Attribute
	for i := range attributes {
		if attributes[i] == nil {
			continue
		}
		x = append(x, newAttribute(*attributes[i]))
	}
	return x
}

func newAttribute(attribute mfd.Attribute) Attribute {
	return Attribute{
		Name:       attribute.Name,
		DBName:     attribute.DBName,
		IsArray:    attribute.IsArray,
		DBType:     attribute.DBType,
		GoType:     attribute.GoType,
		PrimaryKey: attribute.PrimaryKey,
		ForeignKey: attribute.ForeignKey,
		Nullable:   attribute.Nullable(),
		Addable:    attribute.IsAddable(),
		Updatable:  attribute.IsUpdatable(),
		Min:        &attribute.Min, //todo: allow to be nullable
		Max:        &attribute.Max, //todo: allow to be nullable
		Default:    attribute.Default,
	}
}

func newSearches(searches mfd.Searches) []Search {
	var x []Search
	for i := range searches {
		if searches[i] == nil {
			continue
		}
		x = append(x, newSearch(*searches[i]))
	}
	return x
}

func newSearch(search mfd.Search) Search {
	return Search{
		Name:       search.Name,
		AttrName:   search.AttrName,
		SearchType: search.SearchType,
	}
}

func newCustomTypes(types mfd.CustomTypes) []CustomType {
	var x []CustomType
	for i := range types {
		x = append(x, CustomType{
			DBType:   types[i].DBType,
			GoImport: types[i].GoImport,
			GoType:   types[i].GoType,
		})
	}
	return x
}
