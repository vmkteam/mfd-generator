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

// ValidateNamespaceData stores namespace info for template
type ValidateNamespaceData struct {
	GeneratorVersion string
	Package          string

	HasImports bool
	Imports    []string

	Entities []ValidateEntityData
}

// PackValidateNamespace packs mfd namespace to validate template data
func PackValidateNamespace(namespaces []*mfd.Namespace, options Options) ValidateNamespaceData {
	imports := util.NewSet()

	var models []ValidateEntityData
	for _, namespace := range namespaces {
		for _, entity := range namespace.Entities {
			mdl := PackValidateEntity(*entity, options)
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

	return ValidateNamespaceData{
		GeneratorVersion: mfd.Version,
		Package:          options.Package,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Entities: models,
	}
}

// ValidateEntityData stores entity info for template
type ValidateEntityData struct {
	// using model template as base because validate depends on it
	EntityData

	Columns []ValidateAttributeData
	Imports []string
}

// PackValidateEntity packs mfd entity to template data
func PackValidateEntity(entity mfd.Entity, options Options) ValidateEntityData {
	imports := mfd.NewSet()

	columns := make([]ValidateAttributeData, 0, len(entity.Attributes))
	for _, attribute := range entity.Attributes {
		// if field can be validated
		if !isValidatable(*attribute) {
			continue
		}

		tmpl := PackValidateAttribute(entity, *attribute, options)

		columns = append(columns, tmpl)
		if tmpl.Import != "" {
			imports.Add(tmpl.Import)
		}
	}

	return ValidateEntityData{
		// base template entity
		EntityData: PackEntity(entity, options),

		Columns: columns,
		Imports: imports.Elements(),
	}
}

// ValidateAttributeData stores attribute info for validate template
type ValidateAttributeData struct {
	// using model template as base because validate depends on it
	AttributeData

	Check string

	Import string
}

// PackValidateAttribute packs mfd attribute to validate template data
func PackValidateAttribute(entity mfd.Entity, attribute mfd.Attribute, options Options) ValidateAttributeData {
	tmpl := ValidateAttributeData{
		// base template column
		AttributeData: PackAttribute(entity, attribute, options),

		Check: check(attribute),
	}

	if tmpl.Check == PLen || tmpl.Check == Len {
		tmpl.Import = "unicode/utf8"
	}

	return tmpl
}

// isValidatable checks if field can be validated
func isValidatable(attribute mfd.Attribute) bool {
	// do not validate PK
	if attribute.PrimaryKey {
		return false
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
