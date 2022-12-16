package xmlvt

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	nssFlag      = "namespaces"
	entitiesFlag = "entities"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml-vt", "Create vt xml from mfd", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
}

// New creates generator
func New() *Generator {
	return &Generator{}
}

// AddFlags adds flags to command
func (g *Generator) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(nssFlag, "n", []string{}, "namespaces to generate, must be in mfd file. separate by comma")
	flags.StringSliceP(entitiesFlag, "e", []string{}, "entities to generate, must be in vt.xml file. separate by comma")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) error {
	var err error

	flags := command.Flags()

	if g.options.MFDPath, err = flags.GetString(mfdFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nssFlag); err != nil {
		return err
	}

	if g.options.Entities, err = flags.GetStringSlice(entitiesFlag); err != nil {
		return err
	}

	return nil
}

// Generate runs generator
func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, false, 0)
	if err != nil {
		return err
	}

	// validate names
	if err := project.ValidateNames(); err != nil {
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

		entityNames := ns.EntityNames()
		if len(g.options.Entities) != 0 {
			entityNames = g.options.Entities
		}

		for _, name := range entityNames {
			entity := ns.Entity(name)
			if entity == nil {
				return fmt.Errorf("entity %s to generate from not found in project", name)
			}

			exitsting := project.VTEntity(name)

			project.AddVTEntity(namespace, PackVTEntity(entity, exitsting))
		}
	}

	// saving vt xml
	return mfd.SaveProjectVT(g.options.MFDPath, project)
}
