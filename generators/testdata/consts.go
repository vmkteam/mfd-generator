package testdata

import "path/filepath"

const (
	DirTestdata = "testdata"
	DirParent   = ".."

	FilenameMFD   = "newsportal.mfd"
	FilenameXML   = "portal.xml"
	FilenameVTXML = "portal.vt.xml"

	PackageDB         = "db"
	PackageVT         = "vt"
	PackageVTTemplate = "vt-template"
	PackageVTUpdated  = "vt-updated"
	PackageDBTest     = "test"

	PrefixAll         = "all"
	PrefixEntity      = "entities"
	PrefixComposition = "composition"
	PrefixVue3        = "vue3"
)

var (
	PathActual   = filepath.Join(DirParent, DirTestdata, "actual")
	PathExpected = filepath.Join(DirParent, DirTestdata, "expected")

	PathActualMFD                     = filepath.Join(PathActual, FilenameMFD)
	PathExpectedMFD                   = filepath.Join(PathExpected, FilenameMFD)
	PathActualDB                      = filepath.Join(PathActual, PackageDB)
	PathExpectedDB                    = filepath.Join(PathExpected, PackageDB)
	PathActualVT                      = filepath.Join(PathActual, PackageVT)
	PathUpdatedVT                     = filepath.Join(PathExpected, PackageVTUpdated)
	PathExpectedVT                    = filepath.Join(PathExpected, PackageVT)
	PathActualVTTemplateAll           = filepath.Join(PathActual, PackageVTTemplate, PrefixAll)
	PathExpectedVTTemplateAll         = filepath.Join(PathExpected, PackageVTTemplate, PrefixAll)
	PathActualVTTemplateEntity        = filepath.Join(PathActual, PackageVTTemplate, PrefixEntity)
	PathExpectedVTTemplateEntity      = filepath.Join(PathExpected, PackageVTTemplate, PrefixEntity)
	PathActualVTTemplateComposition   = filepath.Join(PathActual, PackageVTTemplate, PrefixComposition)
	PathExpectedVTTemplateComposition = filepath.Join(PathExpected, PackageVTTemplate, PrefixComposition)
	PathActualVTTemplateVue3          = filepath.Join(PathActual, PackageVTTemplate, PrefixVue3)
	PathExpectedVTTemplateVue3        = filepath.Join(PathExpected, PackageVTTemplate, PrefixVue3)
	PathActualDBTest                  = filepath.Join(PathActual, PackageDB, PackageDBTest)
	PathExpectedDBTest                = filepath.Join(PathExpected, PackageDB, PackageDBTest)
)
