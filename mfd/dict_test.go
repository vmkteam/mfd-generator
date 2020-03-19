package mfd

import "testing"

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
