package vt

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

	// Package sets package name for model
	Package string

	// ModelPackage sets package for model files
	ModelPackage string

	// ModelPackage sets package for embedlog
	EmbedLogPackage string

	// Namespaces to generate
	Namespaces []string

	// Entities to generate
	Entities []string

	// go-pg version
	GoPGVer int

	// custom templates
	ModelTemplatePath     string
	ConverterTemplatePath string
	ServiceTemplatePath   string
	ServerTemplatePath    string

	// custom types
	CustomTypes mfd.CustomTypes
}

// Def fills default values of an options
func (o *Options) Def() {
	if strings.Trim(o.Package, " ") == "" {
		o.Package = util.DefaultPackage
	}

	if o.CustomTypes == nil {
		o.CustomTypes = mfd.CustomTypes{}
	}
}
