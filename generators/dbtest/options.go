package dbtest

import (
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// Options stores generator options
type Options struct {
	// Output file path
	Output string

	// MFDPath stores path for mfd project
	MFDPath string

	// Package sets package name for
	Package string

	// DBPackage sets package name for dir with test helpers
	DBPackage string

	// GoPGVer sets package for model files
	GoPGVer int

	// ProjectName to generate a connection to DB by project name
	ProjectName string

	// Namespaces to generate
	Namespaces []string

	// Entities to generate
	Entities []string

	EntitiesByNamespace map[string][]string

	// Force Replaces existing functions
	Force bool

	// custom types
	CustomTypes mfd.CustomTypes
}

// Def fills default values of an options
func (o *Options) Def() {
	if strings.Trim(o.Package, " ") == "" {
		o.Package = util.DefaultPackage
	}

	o.CustomTypes = mfd.CustomTypes{}
	o.EntitiesByNamespace = make(map[string][]string, len(o.Namespaces))
}
