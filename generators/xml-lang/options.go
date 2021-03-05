package xmllang

type Options struct {
	// MFDPath stores path for mfd project
	MFDPath string

	// Languages to generate
	Languages []string

	// Namespaces to generate
	Namespaces []string

	// Entities to generate
	Entities []string
}
