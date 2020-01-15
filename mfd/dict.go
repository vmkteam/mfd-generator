package mfd

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
		"password":    "Пароль",
		"alias":       "Алиас",
		"status":      "Статус",
		"statusId":    "Статус",

		"createdAt":  "Создано",
		"modifiedAt": "Изменено",
		"deletedAt":  "Удалено",

		"actions": "Действия",
	},
}

func Translate(lang, key string) string {
	if _, ok := presetsTranslations[lang]; !ok {
		return ""
	}

	return presetsTranslations[lang][key]
}
