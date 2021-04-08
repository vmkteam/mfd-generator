package api

import (
	"encoding/xml"
	"fmt"
	"log"
	"path/filepath"

	xmlGen "github.com/vmkteam/mfd-generator/generators/xml"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/lib"
	"github.com/vmkteam/zenrpc"
)

const DefaultGoPGVer = mfd.GoPG10

type XMLResponse struct {
	Filename string `json:"filename"`
	XML      string `json:"xml"`
}

func NewXMLResponse(filePath string, v interface{}) (*XMLResponse, error) {
	bytes, err := xml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("xml marshall error: %w", err)
	}

	return &XMLResponse{
		Filename: filePath,
		XML:      string(bytes),
	}, nil
}

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

	entities, err := genna.Read([]string{"*"}, false, false, DefaultGoPGVer, nil)
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
//zenrps:return		xml contents of mfd file
func (s *XMLService) LoadProject(filePath string) (*XMLResponse, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	return NewXMLResponse(filePath, project)
}

// Creates project at filepath location
//zenrpc:filePath	the path to mfd file
//zenrps:return		xml contents of mfd file
func (s *XMLService) CreateProject(filePath string) (*XMLResponse, error) {
	project := mfd.NewProject(filepath.Base(filePath), DefaultGoPGVer)

	err := mfd.SaveMFD(filePath, project)
	if err != nil {
		return nil, err
	}

	return NewXMLResponse(filePath, project)
}

// Saves project at filepath location
//zenrpc:filePath	the path to mfd file
//zenrpc:contents	the xml contents
//zenrps:return		saved xml contents of mfd file
func (s *XMLService) SaveProject(filePath, contents string) (*XMLResponse, error) {
	project := &mfd.Project{}
	if err := xml.Unmarshal([]byte(contents), project); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	err := mfd.SaveMFD(filePath, project)
	if err != nil {
		return nil, err
	}

	return NewXMLResponse(filePath, project)
}

// Gets xml for selected table
//zenrpc:filePath	the path to mfd file
//zenrpc:url		the connection string to postgresql database
//zenrpc:table		selected table name
//zenrpc:namespace	namespace of the created entity
//zenrpc:goPGVer	version of go-pg lib (8,9,10)
//zenrpc:types		custom types mapping
//zenrps:return		xml contents of mfd file
func (s *XMLService) GenerateEntity(filePath, url, table, namespace string) (*XMLResponse, error) {
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

		return NewXMLResponse(filePath, entity)
	}

	return nil, fmt.Errorf("table not found in database")
}

// Gets xml for selected entity in project file
//zenrpc:filePath	the path to mfd file
//zenrpc:namespace	namespace of the entity
//zenrpc:entity 	then name of the entity
//zenrps:return		xml contents of mfd file
func (s *XMLService) LoadEntity(filePath, namespace, entity string) (*XMLResponse, error) {
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

	return NewXMLResponse(filePath, entity)
}

// Gets xml for selected entity in project file
//zenrpc:filePath	the path to mfd file
//zenrpc:contents	xml contents of the entity
//zenrps:return		saved xml contents of the entity
func (s *XMLService) SaveEntity(filePath, contents string) (*XMLResponse, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	entity := &mfd.Entity{}
	if err := xml.Unmarshal([]byte(contents), entity); err != nil {
		return nil, err
	}

	ns := project.Namespace(entity.Namespace)
	if ns == nil {
		ns = project.AddNamespace(entity.Namespace)
	}

	project.AddEntity(ns.Name, entity)

	return NewXMLResponse(filePath, entity)
}

//go:generate zenrpc
