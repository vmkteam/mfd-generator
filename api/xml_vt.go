package api

import (
	"fmt"

	xmlVtGen "github.com/vmkteam/mfd-generator/generators/xml-vt"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/semrush/zenrpc/v2"
)

type XMLVTService struct {
	zenrpc.Service
}

func NewXMLVTService() *XMLVTService {
	return &XMLVTService{}
}

// Gets xml for selected table
//zenrpc:filePath	the path to mfd file
//zenrpc:entity		base entity from namespace.xml
//zenrpc:namespace	namespace of the base entity
//zenrpc:return		VTEntity
func (s *XMLVTService) GenerateEntity(filePath, namespace, entity string) (*mfd.VTEntity, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	ns := project.Namespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	base := ns.Entity(entity)
	if base == nil {
		return nil, fmt.Errorf("table not found in database")
	}

	existing := project.VTEntity(entity)
	vtEntity := xmlVtGen.PackVTEntity(base, existing)

	return vtEntity, nil
}

// Gets xml for selected entity in project file
//zenrpc:filePath	the path to mfd file
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity 	the name of the vt entity
//zenrpc:return		VTEntity
func (s *XMLVTService) LoadEntity(filePath, namespace, entity string) (*mfd.VTEntity, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	ns := project.VTNamespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	ent := ns.VTEntity(entity)
	if ent == nil {
		return nil, fmt.Errorf("entity %s not found", entity)
	}

	return ent, nil
}

// Gets xml for selected entity in project file
//zenrpc:filePath	the path to mfd file
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity		vt entity information
//zenrpc:return		true on success
func (s *XMLVTService) SaveEntity(filePath string, namespace string, entity *mfd.VTEntity) (bool, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return false, err
	}

	ns := project.VTNamespace(namespace)
	if ns == nil {
		ns = project.AddVTNamespace(namespace)
	}

	project.AddVTEntity(ns.Name, entity)
	project.UpdateLinks()

	err = mfd.SaveMFD(filePath, project)
	if err != nil {
		return false, err
	}

	if err := mfd.SaveProjectVT(filePath, project); err != nil {
		return false, err
	}

	return true, nil
}

//go:generate zenrpc
