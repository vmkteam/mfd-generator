package dbtest

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"
)

// NamespaceData stores namespace info for template
type NamespaceData struct {
	Package   string
	DBPackage string

	ProjectName string

	GoPGVer string
}

// PackNamespace packs mfd namespace to template data
func PackNamespace(options Options) NamespaceData {
	goPGVer := ""
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return NamespaceData{
		Package:     options.Package,
		DBPackage:   options.DBPackage,
		ProjectName: options.ProjectName,
		GoPGVer:     goPGVer,
	}
}
