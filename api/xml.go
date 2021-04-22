package api

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/generators/xml"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/semrush/zenrpc/v2"
)

const DefaultGoPGVer = mfd.GoPG10

type XMLService struct {
	*Store

	zenrpc.Service
}

type NSMapping struct {
	Table     string `json:"table"`
	Namespace string `json:"namespace"`
}

func NewXMLService(store *Store) *XMLService {
	return &XMLService{
		Store: store,
	}
}

// Saves project at filepath location
//zenrpc:return		NSMapping
func (s *XMLService) NSMapping() ([]NSMapping, error) {
	index := map[string]struct{}{}
	var result []NSMapping

	for _, ns := range s.CurrentProject.Namespaces {
		for _, entity := range ns.Entities {
			if _, ok := index[entity.Table]; !ok {
				index[entity.Table] = struct{}{}
				result = append(result, NSMapping{Table: entity.Table, Namespace: ns.Name})
			}
		}
	}

	return result, nil
}

// Gets xml for selected table
//zenrpc:table		selected table name
//zenrpc:namespace	namespace of the new entity
//zenrpc:return		Entity
func (s *XMLService) GenerateEntity(table, namespace string) (*mfd.Entity, error) {
	entities, err := s.Genna.Read([]string{table}, false, false, s.CurrentProject.GoPGVer, s.CurrentProject.CustomTypeMapping())
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		exiting := s.CurrentProject.Entity(entity.GoName)

		// adding to project
		entity := xml.PackEntity(namespace, entity, exiting, s.CurrentProject.CustomTypes)

		return entity, nil
	}

	return nil, fmt.Errorf("table not found in database")
}

// Gets xml for selected entity in project file
//zenrpc:namespace	namespace of the entity
//zenrpc:entity 	the name of the entity
//zenrpc:return		Entity
func (s *XMLService) LoadEntity(namespace, entity string) (*mfd.Entity, error) {
	ns := s.CurrentProject.Namespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	ent := ns.Entity(entity)
	if ent == nil {
		return nil, fmt.Errorf("entity %s not found", entity)
	}

	return ent, nil
}

// Gets xml for selected entity in project file
//zenrpc:entity	Entity
func (s *XMLService) SaveEntity(entity *mfd.Entity) error {
	ns := s.CurrentProject.Namespace(entity.Namespace)
	if ns == nil {
		ns = s.CurrentProject.AddNamespace(entity.Namespace)
	}

	s.CurrentProject.AddEntity(ns.Name, entity)
	s.CurrentProject.UpdateLinks()

	return nil
}
