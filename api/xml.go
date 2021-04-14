package api

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	xmlGen "github.com/vmkteam/mfd-generator/generators/xml"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/lib"
	"github.com/vmkteam/zenrpc"
)

const DefaultGoPGVer = mfd.GoPG10

type XMLService struct {
	zenrpc.Service
}

func NewXMLService() *XMLService {
	return &XMLService{}
}

// Gets all tables from database
//zenrpc:url	the connection string to pg database
//zenrps:return	list of tables
func (s *XMLService) Tables(url string) ([]string, error) {
	var logger *log.Logger

	genna := genna.New(url, logger)
	if err := genna.Connect(); err != nil {
		return nil, err
	}

	schemas, err := genna.Store.Schemas()
	if err != nil {
		return nil, err
	}

	var filter []string
	for _, schema := range schemas {
		if strings.HasPrefix(schema, "pg_") || schema == "information_schema" {
			continue
		}
		filter = append(filter, fmt.Sprintf("%s.*", schema))
	}

	entities, err := genna.Read(filter, false, false, DefaultGoPGVer, nil)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(entities))
	for i, entity := range entities {
		result[i] = entity.PGFullName
	}

	return result, nil
}

// Loads project from file
//zenrpc:filePath	the path to mfd file
//zenrps:return		project information
func (s *XMLService) LoadProject(filePath string) (*mfd.Project, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// Creates project at filepath location
//zenrpc:filePath	the path to mfd file
//zenrps:return		project information
func (s *XMLService) CreateProject(filePath string) (*mfd.Project, error) {
	project := mfd.NewProject(filepath.Base(filePath), DefaultGoPGVer)

	err := mfd.SaveMFD(filePath, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// Saves project at filepath location
//zenrpc:filePath	the path to mfd file
//zenrpc:project	project information
func (s *XMLService) SaveProject(filePath string, project mfd.Project) (bool, error) {
	original, err := mfd.LoadProject(filePath, true, project.GoPGVer)
	if err != nil {
		return false, err
	}

	project.XMLName = original.XMLName
	project.XMLxsd = original.XMLxsd
	project.XMLxsi = original.XMLxsi

	err = mfd.SaveMFD(filePath, &project)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Saves project at filepath location
//zenrpc:filePath	the path to mfd file
//zenrps:return		table-namespace mapping
func (s *XMLService) NSMapping(filePath string) (map[string]string, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for _, ns := range project.Namespaces {
		for _, entity := range ns.Entities {
			result[entity.Table] = ns.Name
		}
	}

	return result, nil
}

// Gets xml for selected table
//zenrpc:filePath	the path to mfd file
//zenrpc:url		the connection string to postgresql database
//zenrpc:table		selected table name
//zenrpc:namespace	namespace of the new entity
//zenrps:return		entity information
func (s *XMLService) GenerateEntity(filePath, url, table, namespace string) (*mfd.Entity, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	genna := genna.New(url, nil)

	entities, err := genna.Read([]string{table}, false, false, project.GoPGVer, project.CustomTypeMapping())
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		exiting := project.Entity(entity.GoName)

		// adding to project
		entity := xmlGen.PackEntity(namespace, entity, exiting, project.CustomTypes)

		return entity, nil
	}

	return nil, fmt.Errorf("table not found in database")
}

// Gets xml for selected entity in project file
//zenrpc:filePath	the path to mfd file
//zenrpc:namespace	namespace of the entity
//zenrpc:entity 	the name of the entity
//zenrps:return		entity information
func (s *XMLService) LoadEntity(filePath, namespace, entity string) (*mfd.Entity, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	ns := project.Namespace(namespace)
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
//zenrpc:filePath	the path to mfd file
//zenrpc:contents	xml contents of the entity
func (s *XMLService) SaveEntity(filePath string, entity *mfd.Entity) (bool, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return false, err
	}

	ns := project.Namespace(entity.Namespace)
	if ns == nil {
		ns = project.AddNamespace(entity.Namespace)
	}

	project.AddEntity(ns.Name, entity)
	project.UpdateLinks()

	err = mfd.SaveMFD(filePath, project)
	if err != nil {
		return false, err
	}

	if err := mfd.SaveProjectXML(filePath, project); err != nil {
		return false, err
	}

	return true, nil
}

//go:generate zenrpc
