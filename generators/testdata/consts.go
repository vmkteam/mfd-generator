package testdata

const (
	FilenameMfd   = "newsportal.mfd"
	FilenameXml   = "portal.xml"
	FilenameVtXml = "portal.vt.xml"

	PackageDB         = "db"
	PackageVt         = "vt"
	PackageVtTemplate = "vt-template"

	PathActual                   = "../testdata/actual/"
	PathExpected                 = "../testdata/expected/"
	PathActualMfd                = PathActual + FilenameMfd
	PathExpectedMfd              = PathExpected + FilenameMfd
	PathActualDB                 = PathActual + PackageDB + "/"
	PathExpectedDB               = PathExpected + PackageDB + "/"
	PathActualVt                 = PathActual + PackageVt + "/"
	PathExpectedVt               = PathExpected + PackageVt + "/"
	PathActualVtTemplateAll      = PathActual + PackageVtTemplate + PrefixAll + "/"
	PathExpectedVtTemplateAll    = PathExpected + PackageVtTemplate + PrefixAll + "/"
	PathActualVtTemplateEntity   = PathActual + PackageVtTemplate + PrefixEntity + "/"
	PathExpectedVtTemplateEntity = PathExpected + PackageVtTemplate + PrefixEntity + "/"

	PrefixAll    = "/all"
	PrefixEntity = "/entities"
)
