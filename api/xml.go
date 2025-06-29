package api

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"text/template"

	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/generators/xml"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/vmkteam/zenrpc/v2"
)

const DefaultGoPGVer = mfd.GoPG10

type XMLService struct {
	*Store

	zenrpc.Service
}

func NewXMLService(store *Store) *XMLService {
	return &XMLService{
		Store: store,
	}
}

// GenerateEntity returns entity for selected table.
//
//zenrpc:table		selected table name
//zenrpc:namespace	namespace of the new entity
//zenrpc:return		Entity
func (s XMLService) GenerateEntity(table, namespace string) (*mfd.Entity, error) {
	entities, err := s.Genna.Read([]string{table}, false, false, s.CurrentProject.GoPGVer, s.CurrentProject.CustomTypeMapping())
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		exiting := s.CurrentProject.EntityByTable(entity.PGFullName)

		// adding to project
		entity := xml.PackEntity(namespace, entity, exiting, s.CurrentProject.CustomTypes)

		return entity, nil
	}

	return nil, errors.New("table not found in database")
}

// LoadEntity returns selected entity from project.
//
//zenrpc:namespace	namespace of the entity
//zenrpc:entity 	the name of the entity
//zenrpc:return		Entity
func (s XMLService) LoadEntity(namespace, entity string) (*mfd.Entity, error) {
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

// UpdateEntity saves selected entity in project.
//
//zenrpc:entity	Entity
func (s XMLService) UpdateEntity(entity *mfd.Entity) error {
	ns := s.CurrentProject.Namespace(entity.Namespace)
	if ns == nil {
		ns = s.CurrentProject.AddNamespace(entity.Namespace)
	}

	s.CurrentProject.AddEntity(ns.Name, entity)
	s.CurrentProject.UpdateLinks()

	return nil
}

// GenerateModelCode generates model go code, that represents this entity.
//
//zenrpc:entity Entity
func (s XMLService) GenerateModelCode(entity mfd.Entity) (string, error) {
	ent := model.PackEntity(entity, model.Options{GoPGVer: s.CurrentProject.GoPGVer, CustomTypes: s.CurrentProject.CustomTypes})
	tpl := template.Must(template.New("tmp").Parse(modelTemplate))
	var b bytes.Buffer
	err := tpl.Execute(&b, ent)
	if err != nil {
		return "", err
	}
	res, err := format.Source(b.Bytes())
	return string(res), err
}

// GenerateSearchModelCode generates search go code, that represents this entity.
//
//zenrpc:entity Entity
func (s XMLService) GenerateSearchModelCode(entity mfd.Entity) (string, error) {
	// TODO PackSearchEntity panics
	ent := model.PackEntity(entity, model.Options{GoPGVer: s.CurrentProject.GoPGVer, CustomTypes: s.CurrentProject.CustomTypes})
	tpl := template.Must(template.New("tmp").Parse(searchTemplate))
	var b bytes.Buffer
	err := tpl.Execute(&b, ent)
	if err != nil {
		return "", err
	}
	res, err := format.Source(b.Bytes())
	return string(res), err
}

const modelTemplate = `type {{.Name}} struct {
    tableName struct{} {{.Tag}}
    {{range .Columns}}
    {{.Name}} {{.GoType}} {{.Tag}} {{.Comment}}{{end}}{{if .HasRelations}}
    {{range .Relations}}
    {{.Name}} *{{.Type}} {{.Tag}} {{.Comment}}{{end}}{{end}}
}`

const searchTemplate = `type {{.Name}}Search struct {
	// fixme: wrong search type
    search 

    {{range .Columns}}
    {{.Name}} {{.GoType}}{{end}}
}`
