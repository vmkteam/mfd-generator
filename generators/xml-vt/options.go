package xmlvt

type Options struct {
	// MFDPath stores path for mfd project
	MFDPath string

	// Namespaces to generate
	Namespaces []string

	// Entities to generate
	Entities []string
}
