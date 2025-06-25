package api

import (
	"errors"
	"fmt"

	xmlvt "github.com/vmkteam/mfd-generator/generators/xml-vt"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/vmkteam/zenrpc/v2"
)

type XMLVTService struct {
	*Store

	zenrpc.Service
}

func NewXMLVTService(store *Store) *XMLVTService {
	return &XMLVTService{
		Store: store,
	}
}

// GenerateEntity returns vt entity for selected base entity.
//
//zenrpc:namespace	namespace of the base entity
//zenrpc:entity		base entity from namespace.xml
//zenrpc:return		VTEntity
func (s XMLVTService) GenerateEntity(namespace, entity string) (*mfd.VTEntity, error) {
	ns := s.CurrentProject.Namespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	base := ns.Entity(entity)
	if base == nil {
		return nil, errors.New("table not found in database")
	}

	existing := s.CurrentProject.VTEntity(entity)
	vtEntity := xmlvt.PackVTEntity(base, existing)

	return vtEntity, nil
}

// LoadEntity returns vt entity for selected entity from project.
//
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity 	the name of the vt entity
//zenrpc:return		VTEntity
func (s XMLVTService) LoadEntity(namespace, entity string) (*mfd.VTEntity, error) {
	ns := s.CurrentProject.VTNamespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	ent := ns.VTEntity(entity)
	if ent == nil {
		return nil, fmt.Errorf("entity %s not found", entity)
	}

	return ent, nil
}

// UpdateEntity saves vt entity in project.
//
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity		VTEntity
func (s XMLVTService) UpdateEntity(namespace string, entity *mfd.VTEntity) error {
	ns := s.CurrentProject.VTNamespace(namespace)
	if ns == nil {
		ns = s.CurrentProject.AddVTNamespace(namespace)
	}

	s.CurrentProject.AddVTEntity(ns.Name, entity)
	s.CurrentProject.UpdateLinks()

	return nil
}
