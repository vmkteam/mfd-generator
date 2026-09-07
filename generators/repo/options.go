package repo

import (
	"fmt"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// Mode selects the repository template family.
type Mode string

const (
	ModeLegacy  Mode = "legacy"
	ModeGeneric Mode = "generic"
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

	// Mode selects the generated repository API.
	Mode Mode
}

// Def fills default values of an options
func (o *Options) Def() {
	if strings.Trim(o.Package, " ") == "" {
		o.Package = util.DefaultPackage
	}

	if o.CustomTypes == nil {
		o.CustomTypes = mfd.CustomTypes{}
	}

	if o.Mode == "" {
		o.Mode = ModeLegacy
	}
}

// Validate checks options that affect the generated repository contract.
func (o Options) Validate() error {
	switch o.Mode {
	case ModeLegacy, ModeGeneric:
		return nil
	default:
		return fmt.Errorf("unknown repo mode %q", o.Mode)
	}
}
