package vt

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

	// ModelPackage sets package for model files
	ModelPackage string

	// Namespaces to generate
	Namespaces []string

	// go-pg version
	GoPGVer int
}

// Def fills default values of an options
func (o *Options) Def() {
	if strings.Trim(o.Package, " ") == "" {
		o.Package = util.DefaultPackage
	}
}
