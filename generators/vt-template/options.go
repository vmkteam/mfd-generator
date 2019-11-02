package vttmpl

// Options stores generator options
type Options struct {
	// Output file path
	Output string

	// MFDPath stores path for mfd project
	MFDPath string

	// Namespaces to generate
	Namespaces []string

	ListTemplate    string
	FiltersTemplate string
	FormTemplate    string
}
