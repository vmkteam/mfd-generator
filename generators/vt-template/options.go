package vttmpl

// Options stores generator options
type Options struct {
	// Output file path
	Output string

	// MFDPath stores path for mfd project
	MFDPath string

	// Namespaces to generate
	Namespaces []string

	// Entities to generate
	Entities []string

	// custom templates
	RoutesTemplatePath  string
	ListTemplatePath    string
	FiltersTemplatePath string
	FormTemplatePath    string
}
