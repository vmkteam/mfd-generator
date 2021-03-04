package model

import (
	"strings"

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

	// go-pg version
	GoPGVer int

	// custom templates
	ModelTemplatePath    string
	SearchTemplatePath   string
	ValidateTemplatePath string
}

// Def fills default values of an options
func (o *Options) Def() {
	if strings.Trim(o.Package, " ") == "" {
		o.Package = util.DefaultPackage
	}
}
