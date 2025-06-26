package repo

import (
	"fmt"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag = "mfd"
	pkgFlag = "package"
	nsFlag  = "namespaces"

	repoTemplateFlag = "repo-tmpl"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("repo", "Create repo from xml", New())
}

// Generator represents repo generator
type Generator struct {
	options Options
}

// New creates repo generator
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

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used")

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces to generate. separate by comma\n")

	flags.String(repoTemplateFlag, "", "path to repo custom template\n")
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

	if g.options.Namespaces, err = flags.GetStringSlice(nsFlag); err != nil {
		return err
	}

	if g.options.Package, err = flags.GetString(pkgFlag); err != nil {
		return err
	}

	if g.options.Package == "" {
		g.options.Package = path.Base(g.options.Output)
	}

	if g.options.RepoTemplatePath, err = flags.GetString(repoTemplateFlag); err != nil {
		return err
	}

	g.options.Def()

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

	g.options.GoPGVer = project.GoPGVer
	g.options.CustomTypes = project.CustomTypes

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	repoTemplate, err := mfd.LoadTemplate(g.options.RepoTemplatePath, repoDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load repo template, err=%w", err)
	}

	for _, namespace := range g.options.Namespaces {
		// generating each namespace in separate file
		if ns := project.Namespace(namespace); ns != nil {
			// getting file name without dots
			output := path.Join(g.options.Output, mfd.GoFileName(namespace)+".go")
			data := PackNamespace(ns, g.options)
			if _, err := mfd.FormatAndSave(data, output, repoTemplate, true); err != nil {
				return fmt.Errorf("generate repo %s, err=%w", namespace, err)
			}
		}
	}

	return nil
}
