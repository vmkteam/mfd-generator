package repo

import (
	"github.com/vmkteam/mfd-generator/mfd"
	"path"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

const (
	mfdFlag = "mfd"
	pkgFlag = "package"
	nsFlag  = "ns"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("repo", "Create repo from xml", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
}

// New creates basic generator
func New() *Generator {
	return &Generator{}
}

// AddFlags adds flags to command
func (g *Generator) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.StringP(base.Output, "o", "", "output dir path")
	if err := command.MarkFlagRequired(base.Output); err != nil {
		panic(err)
	}

	flags.StringP(mfdFlag, "m", "", "mfd file")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringP(pkgFlag, "p", "", "package name")

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces")
}

// ReadFlags read flags from command
func (g *Generator) ReadFlags(command *cobra.Command) error {
	var err error

	flags := command.Flags()

	if g.options.Output, err = flags.GetString(base.Output); err != nil {
		return err
	}

	if g.options.MFDPath, err = flags.GetString(mfdFlag); err != nil {
		return err
	}

	if g.options.Package, err = flags.GetString(pkgFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nsFlag); err != nil {
		return err
	}

	g.options.Def()

	return nil
}

// Packer returns packer function for compile entities into package
func (g *Generator) Packer(namespace string) mfd.Packer {
	return func(namespaces mfd.Namespaces) (interface{}, error) {
		return NewTemplatePackage(namespace, namespaces, g.options), nil
	}
}

func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, false)
	if err != nil {
		return err
	}

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	for _, ns := range g.options.Namespaces {
		// generating each namespace in separate file
		if p := project.Namespace(ns); p != nil {
			// getting file name without dots
			output := path.Join(g.options.Output, mfd.GoFileName(ns)+".go")
			if _, err := mfd.PackAndSave(project.Namespaces, output, repoTemplate, g.Packer(ns), true); err != nil {
				return xerrors.Errorf("generate repo %s error: %w", ns, err)
			}
		}
	}

	return nil
}
