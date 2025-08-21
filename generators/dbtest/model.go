package dbtest

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"
)

// FuncRenderData stores data for generating functions template
type FuncRenderData struct {
	Package   string
	DBPackage string

	ProjectName string
	GoPGVer     string
}

// PackFuncRenderData packs mfd namespace to template data
func PackFuncRenderData(options Options) FuncRenderData {
	var goPGVer string
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return FuncRenderData{
		GoPGVer:     goPGVer,
		Package:     options.Package,
		DBPackage:   options.DBPackage,
		ProjectName: options.ProjectName,
	}
}
