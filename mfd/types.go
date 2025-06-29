package mfd

import (
	"fmt"
	"strings"

	"github.com/dizzyfool/genna/model"
)

func MakeSearchType(typ string, searchType SearchType) string {
	if typ[0] == '*' {
		typ = typ[1:]
	}

	switch searchType {
	case SearchArray, SearchNotArray, SearchTypeArrayContained, SearchTypeArrayIntersect:
		if _, ok := IsArray(typ); ok {
			return typ
		}
		return "[]" + typ
	case SearchTypeArrayContains, SearchTypeArrayNotContains:
		if el, ok := IsArray(typ); ok {
			return "*" + el
		}
		return "*" + typ
	case SearchNull, SearchNotNull:
		return "*" + model.TypeBool
	case SearchTypeJsonbPath:
		return "*" + model.TypeString
	}

	if _, ok := IsArray(typ); ok {
		return typ
	}
	return "*" + typ
}

func Element(typ string) (el string) {
	el, _ = IsArray(typ)
	el, _ = IsPointer(el)

	return
}

func IsPointer(typ string) (string, bool) {
	if typ != "" && typ[0] == '*' {
		return typ[1:], true
	}

	return typ, false
}

func IsArray(typ string) (string, bool) {
	if typ != "" && typ[0] == '[' {
		return typ[2:], true
	}

	return typ, false
}

func IsJSON(typ string) bool {
	return strings.Contains(typ, JSONFieldSep)
}

func MakeZeroValue(typ string) string {
	switch typ {
	case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64, model.TypeDuration:
		return "0"
	case model.TypeString:
		return `""`
	case model.TypeBool:
		return "false"
	case model.TypeTime:
		return "time.Time{}"
	case model.TypeIP:
		return "new.IP{}"
	case model.TypeIPNet:
		return "new.IPNet{}"
	}

	return "nil"
}

func MakeJSType(typ string, isArray bool) string {
	if typ[0] == '*' {
		typ = typ[1:]
	}

	var jsType string
	switch typ {
	case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64:
		jsType = "number"
	case model.TypeBool:
		jsType = "boolean"
	default:
		jsType = "string"
	}

	if isArray {
		return fmt.Sprintf("Array<%s>", jsType)
	}

	return jsType
}

func MakeJSZero(typ string, isArray bool) string {
	if typ[0] == '*' {
		typ = typ[1:]
	}

	var jsZero string
	switch typ {
	case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64:
		jsZero = "0"
	case model.TypeBool:
		jsZero = "false"
	default:
		jsZero = `""`
	}

	if isArray {
		return fmt.Sprintf("[%s]", jsZero)
	}

	return jsZero
}

// Import gets import string for template
func Import(attribute *Attribute, goPGVer int, customTypes CustomTypes) string {
	if customTypes != nil {
		if imp, ok := customTypes.GoImport(Element(attribute.GoType), attribute.DBType); ok {
			return imp
		}
	}

	return model.GoImport(attribute.DBType, attribute.Nullable(), false, goPGVer)
}
