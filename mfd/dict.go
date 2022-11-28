package mfd

import (
	"strings"

	"github.com/fatih/camelcase"
	"github.com/jinzhu/inflection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	EnLang = "en"
	RuLang = "ru"
)

var presetsTranslations = map[string]map[string]string{
	RuLang: {
		"add":    "Добавить",
		"edit":   "Редактировать",
		"delete": "Удалить",

		"user":       "Пользователь",
		"users":      "Пользователи",
		"category":   "Категория",
		"categories": "Категории",
		"tag":        "Тег",
		"tags":       "Теги",

		"title":       "Название",
		"description": "Описание",
		"foreword":    "Краткое содержание",
		"content":     "Содержание",
		"email":       "Email",
		"login":       "Логин",
		"image":       "Изображение",
		"password":    "Пароль",
		"alias":       "Системное имя",
		"status":      "Статус",
		"statusId":    "Статус",

		"createdAt":  "Создано",
		"modifiedAt": "Изменено",
		"deletedAt":  "Удалено",

		"actions": "Действия",
	},
	EnLang: {
		"add":    "Add",
		"edit":   "Edit",
		"delete": "Delete",

		"user":       "User",
		"users":      "Users",
		"category":   "Category",
		"categories": "Categories",
		"tag":        "Tag",
		"tags":       "Tags",

		"title":       "Title",
		"description": "Description",
		"foreword":    "Foreword",
		"content":     "Content",
		"email":       "Email",
		"login":       "Login",
		"image":       "Image",
		"password":    "Password",
		"alias":       "Alias",
		"status":      "Status",
		"statusId":    "Status",

		"createdAt":  "Created at",
		"modifiedAt": "Modified at",
		"deletedAt":  "Deleted at",

		"actions": "Actions",
	},
}

func Translate(lang, key string) string {
	// check current lang
	if _, ok := presetsTranslations[lang]; !ok {
		return ""
	}

	// try to find from base translations
	if found, ok := presetsTranslations[lang][key]; ok {
		return found
	}

	c := cases.Title(language.Und, cases.NoLower)

	// try to convert camelcase for EN
	if lang == EnLang && key != "" {
		if strings.HasSuffix(key, "Ids") {
			key = strings.TrimSuffix(key, "Ids")
			key = inflection.Plural(key)
		} else {
			key = strings.TrimSuffix(key, "Id")
		}

		splitted := camelcase.Split(key)
		splitted[0] = c.String(splitted[0]) // uppercase first letter

		return strings.Join(splitted, " ")
	}

	return ""
}
