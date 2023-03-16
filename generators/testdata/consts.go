package testdata

const (
	PackageDB         = "db"
	PackageVt         = "vt"
	PackageVtTemplate = "vt-template"
	FilenameMfd       = "newsportal.mfd"
	PathActual        = "../testdata/necessary-content/actual/"
	PathExpected      = "../testdata/necessary-content/expected/"

	AllPrefix    = "/all"
	EntityPrefix = "/entities"

	PathExpectedMfd              = PathExpected + FilenameMfd
	PathActualMfd                = PathActual + FilenameMfd
	PathActualDB                 = PathActual + PackageDB + "/"
	PathExpectedDB               = PathExpected + PackageDB + "/"
	PathActualVt                 = PathActual + PackageVt + "/"
	PathExpectedVt               = PathExpected + PackageVt + "/"
	PathActualVtTemplateAll      = PathActual + PackageVtTemplate + AllPrefix + "/"
	PathExpectedVtTemplateAll    = PathExpected + PackageVtTemplate + AllPrefix + "/"
	PathActualVtTemplateEntity   = PathActual + PackageVtTemplate + EntityPrefix + "/"
	PathExpectedVtTemplateEntity = PathExpected + PackageVtTemplate + EntityPrefix + "/"
)
