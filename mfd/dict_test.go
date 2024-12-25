package mfd

import (
	"encoding/xml"
	"testing"
)

func TestTranslate(t *testing.T) {
	var testCases = []struct {
		ln  string
		in  string
		out string
	}{
		{RuLang, "", ""},
		{RuLang, "users", "Пользователи"},
		{EnLang, "userId", "User"},
		{EnLang, "user", "User"},
		{EnLang, "notExists", "Not Exists"},
		{EnLang, "tagIds", "Tags"},
		{EnLang, "testId", "Test"},
		{EnLang, "asd", "Asd"},
	}
	for _, tt := range testCases {
		s := Translate(tt.ln, tt.in)
		if s != tt.out {
			t.Errorf("%v: got %q, want %q", tt.ln, s, tt.out)
		}
	}
}

func TestAddedCustomTranslations(t *testing.T) {
	var testCases = []struct {
		name string
		in   string
		dict *Dict
		want string
	}{
		{
			"with updated word",
			"user",
			&Dict{
				Entries: []Entry{
					{XMLName: xml.Name{Local: "user"}, Value: "Пользователь (Обновленный)"},
				},
			},
			"Пользователь (Обновленный)",
		},
		{
			"with new word",
			"newKey",
			&Dict{
				Entries: []Entry{
					{XMLName: xml.Name{Local: "newKey"}, Value: "Изображение (Обновленное)"},
				},
			},
			"Изображение (Обновленное)",
		},
		{
			"with empty dict",
			"",
			&Dict{
				Entries: []Entry{},
			},
			"",
		},
		{
			"with nil dict",
			"",
			nil,
			"",
		},
	}
	for _, tt := range testCases {
		old := Translate(RuLang, tt.in)
		AddedCustomTranslations(tt.dict)
		updated := Translate(RuLang, tt.in)
		if updated != tt.want {
			t.Errorf("%v: got %q, want %q", tt.in, old, tt.want)
		}
	}
}
