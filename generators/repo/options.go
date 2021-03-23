package repo

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

	// Namespaces to generate
	Namespaces []string

	// go-pg version
	GoPGVer int

	// custom templates
	RepoTemplatePath string

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
