package api

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/generators/xml-lang"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/semrush/zenrpc/v2"
)

type XMLLangService struct {
	zenrpc.Service
}

func NewXMLLangService() *XMLLangService {
	return &XMLLangService{}
}

// Loads full translation of project
//zenrpc:filePath	the path to mfd file
//zenrpc:language   language
//zenrpc:return		Translation
func (s *XMLLangService) LoadTranslation(filePath, language string) (*mfd.Translation, error) {
	_, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	translations, err := mfd.LoadTranslations(filePath, []string{language})
	if err != nil {
		return nil, err
	}

	if translation, ok := translations[language]; ok {
		return &translation, nil
	}

	return nil, nil
}

// Translates entity
//zenrpc:filePath	the path to mfd file
//zenrpc:namespace	namespace of the vt entity
//zenrpc:entity		vt entity from vt.xml
//zenrpc:language   language
//zenrpc:return		TranslationEntity
func (s *XMLLangService) TranslateEntity(filePath, namespace, entity, language string) (*mfd.TranslationEntity, error) {
	project, err := mfd.LoadProject(filePath, false, DefaultGoPGVer)
	if err != nil {
		return nil, err
	}

	translations, err := mfd.LoadTranslations(filePath, []string{language})
	if err != nil {
		return nil, err
	}

	translation, ok := translations[language]
	if !ok {
		translation = mfd.Translation{}
	}

	ns := project.VTNamespace(namespace)
	if ns == nil {
		return nil, fmt.Errorf("vt namespace %s is not found in project", namespace)
	}

	translation = xmllang.Translate(ns, translation, []string{entity}, language)

	return translation.Entity(namespace, entity), nil
}
