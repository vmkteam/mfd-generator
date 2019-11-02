package model

import (
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

const (
	// Nil is nil check types
	Nil = "nil"
	// Zero is 0 check types
	Zero = "zero"
	// PZero is 0 check types for pointers
	PZero = "pzero"
	// Len is length check types
	Len = "len"
	// PLen is length check types for pointers
	PLen = "plen"
	// Enum is allowed values check types
	Enum = "enum"
)

// TemplatePackage stores package info
type ValidateTemplatePackage struct {
	Package string

	HasImports bool
	Imports    []string

	Entities []ValidateTemplateEntity
}

// NewTemplatePackage creates a package for template
func NewValidateTemplatePackage(namespaces mfd.Namespaces, options Options) ValidateTemplatePackage {
	imports := util.NewSet()

	var models []ValidateTemplateEntity
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			mdl := NewValidateTemplateEntity(*entity, options)
			// if there is nothing to validate - skip
			if len(mdl.Columns) == 0 {
				continue
			}

			for _, imp := range mdl.Imports {
				imports.Add(imp)
			}

			models = append(models, mdl)
		}
	}

	return ValidateTemplatePackage{
		Package: options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}
}

// TemplateEntity stores struct info
type ValidateTemplateEntity struct {
	// using model template as base because validate depends on it
	TemplateEntity

	Columns []ValidateTemplateColumn
	Imports []string
}

// NewTemplateEntity creates an entity for template
func NewValidateTemplateEntity(entity mfd.Entity, options Options) ValidateTemplateEntity {
	imports := mfd.NewSet()

	var columns []ValidateTemplateColumn
	for _, attribute := range entity.Attributes {
		// if field can be validated
		if !isValidatable(*attribute) {
			continue
		}

		tmpl := NewValidateTemplateColumn(entity, *attribute, options)

		columns = append(columns, tmpl)
		if tmpl.Import != "" {
			imports.Add(tmpl.Import)
		}
	}

	return ValidateTemplateEntity{
		// base template entity
		TemplateEntity: NewTemplateEntity(entity, options),

		Columns: columns,
		Imports: imports.Elements(),
	}
}

// TemplateColumn stores column info
type ValidateTemplateColumn struct {
	// using model template as base because validate depends on it
	TemplateColumn

	Check string

	Import string
}

// NewTemplateColumn creates a column for template
func NewValidateTemplateColumn(entity mfd.Entity, attribute mfd.Attribute, options Options) ValidateTemplateColumn {
	tmpl := ValidateTemplateColumn{
		// base template column
		TemplateColumn: NewTemplateColumn(entity, attribute, options),

		Check: check(attribute),
	}

	if tmpl.Check == PLen || tmpl.Check == Len {
		tmpl.Import = "unicode/utf8"
	}

	return tmpl
}

// isValidatable checks if field can be validated
func isValidatable(attribute mfd.Attribute) bool {
	// validate FK
	if attribute.PrimaryKey {
		return true
	}

	// validate complex types
	if (attribute.IsArray || attribute.IsJSON() || attribute.IsMap()) && attribute.Nullable() {
		return true
	}

	// validate strings len
	if attribute.IsString() && attribute.Max > 0 {
		return true
	}

	return false
}

// check return check type for validation
func check(attribute mfd.Attribute) string {
	if !isValidatable(attribute) {
		return ""
	}

	// if array/hstore - validate for nil
	if attribute.IsArray || attribute.IsMap() {
		return Nil
	}

	// if pk & int - validate for 0
	if attribute.PrimaryKey && attribute.IsInteger() {
		return Zero
	}

	// if fk & int - validate for 0
	if attribute.ForeignKey != "" && attribute.IsInteger() {
		if attribute.Nullable() {
			return PZero
		}
		return Zero
	}

	// validate for string max len
	if attribute.Max > 0 && attribute.IsString() {
		if attribute.Nullable() {
			return PLen
		}
		return Len
	}

	return ""
}
