package api

import (
	"fmt"
	"log"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	genna "github.com/dizzyfool/genna/lib"
	"github.com/vmkteam/zenrpc/v2"
)

type ProjectService struct {
	*Store

	zenrpc.Service
}

func NewProjectService(store *Store) *ProjectService {
	return &ProjectService{
		Store: store,
	}
}

// Open loads project from file.
//
//zenrpc:filePath		the path to mfd file
//zenrpc:connection 	connection string
//zenrpc:return 		Project
func (s ProjectService) Open(filePath, connection string) (*mfd.Project, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	var logger *log.Logger

	genna := genna.New(connection, logger)
	if err := genna.Connect(); err != nil {
		return nil, err
	}

	s.CurrentProject = project
	s.CurrentFile = filePath
	s.Genna = &genna

	return project, nil
}

// Current returns currently opened project.
//
//zenrpc:return 		Project
func (s ProjectService) Current() (*mfd.Project, error) {
	return s.CurrentProject, nil
}

// Update updates project in memory.
//
//zenrpc:project	Project
func (s ProjectService) Update(project mfd.Project) error {
	s.CurrentProject = &project
	s.CurrentProject.UpdateByNSMapping()

	return nil
}

// Save saves project from memory to disk.
func (s ProjectService) Save() error {
	if err := mfd.SaveMFD(s.CurrentFile, s.CurrentProject); err != nil {
		return err
	}

	if err := mfd.SaveProjectXML(s.CurrentFile, s.CurrentProject); err != nil {
		return err
	}

	if err := mfd.SaveProjectVT(s.CurrentFile, s.CurrentProject); err != nil {
		return err
	}

	// TODO Save translation

	return nil
}

// Tables returns all tables from database.
//
//zenrpc:url	the connection string to pg database
//zenrpc:return	list of tables
func (s ProjectService) Tables() ([]string, error) {
	schemas, err := s.Genna.Store.Schemas()
	if err != nil {
		return nil, err
	}

	filter := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		if strings.HasPrefix(schema, "pg_") || schema == "information_schema" {
			continue
		}
		filter = append(filter, fmt.Sprintf("%s.*", schema))
	}

	entities, err := s.Genna.Read(filter, false, false, DefaultGoPGVer, nil)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(entities))
	for i, entity := range entities {
		result[i] = entity.PGFullName
	}

	return result, nil
}
