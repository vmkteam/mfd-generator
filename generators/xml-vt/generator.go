package xmlvt

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag = "mfd"
	nsFlag  = "ns"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml-vt", "Create vt xml from database", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
	verbose bool
	base    base.Generator
}

// New creates generator
func New() *Generator {
	return &Generator{}
}

// AddFlags adds flags to command
func (g *Generator) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.StringP(mfdFlag, "m", "", "mfd file")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) error {
	var err error

	flags := command.Flags()

	if g.options.MFDPath, err = flags.GetString(mfdFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nsFlag); err != nil {
		return err
	}

	return nil
}

// Generate runs generator
func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, true)
	if err != nil {
		return err
	}

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	// adding vt entities to project
	for _, namespace := range g.options.Namespaces {
		ns := project.Namespace(namespace)
		if ns == nil {
			return fmt.Errorf("namespace %s not found", namespace)
		}

		for _, entity := range ns.Entities {
			project.AddVTEntity(namespace, PackVTEntity(entity))
		}
	}

	// saving vt xml
	return mfd.SaveProjectVT(g.options.MFDPath, project)
}
