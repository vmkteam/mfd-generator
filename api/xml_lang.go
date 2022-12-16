package api

import (
	"fmt"

	xmllang "github.com/vmkteam/mfd-generator/generators/xml-lang"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/vmkteam/zenrpc/v2"
)

type XMLLangService struct {
	*Store

	zenrpc.Service
}

func NewXMLLangService(store *Store) *XMLLangService {
	return &XMLLangService{
		Store: store,
	}
}

// LoadTranslation loads full translation of project.
//
//zenrpc:language   language
//zenrpc:return		Translation
func (s XMLLangService) LoadTranslation(language string) (*mfd.Translation, error) {
	translations, err := mfd.LoadTranslations(s.CurrentFile, []string{language})
	if err != nil {
		return nil, err
	}

	if translation, ok := translations[language]; ok {
		return &translation, nil
	}

	return nil, nil
}

// TranslateEntity translates entity.
//
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity		vt entity from vt.xml
//zenrpc:language   language
//zenrpc:return		TranslationEntity
func (s XMLLangService) TranslateEntity(namespace, entity, language string) (*mfd.TranslationEntity, error) {
	translations, err := mfd.LoadTranslations(s.CurrentFile, []string{language})
	if err != nil {
		return nil, err
	}

	translation, ok := translations[language]
	if !ok {
		translation = mfd.Translation{}
	}

	ns := s.CurrentProject.VTNamespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("vt namespace %s is not found in project", namespace)
	}

	translation = xmllang.Translate(ns, translation, []string{entity}, language)

	return translation.Entity(namespace, entity), nil
}
