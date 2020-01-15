package vt

import (
	"bytes"
	"fmt"
	"html/template"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

const (
	mfdFlag      = "mfd"
	pkgFlag      = "package"
	modelPkgFlag = "model-pkg"
	nsFlag       = "ns"
)

// CreateCommand creates generator command
func CreateCommand(logger *zap.Logger) *cobra.Command {
	return base.CreateCommand("vt", "Create vt from xml", New(logger))
}

// Generator represents mfd generator
type Generator struct {
	logger  *zap.Logger
	options Options
}

// New creates basic generator
func New(logger *zap.Logger) *Generator {
	return &Generator{
		logger: logger,
	}
}

// Logger gets logger
func (g *Generator) Logger() *zap.Logger {
	return g.logger
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

	flags.StringP(modelPkgFlag, "x", "", "package with model files")
	if err := command.MarkFlagRequired(modelPkgFlag); err != nil {
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

	if g.options.ModelPackage, err = flags.GetString(modelPkgFlag); err != nil {
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
func (g *Generator) ModelPacker() mfd.Packer {
	return func(namespaces mfd.Namespaces) (interface{}, error) {
		return NewTemplatePackage(namespaces, g.options)
	}
}

// Packer returns packer function for compile entities into package
func (g *Generator) ServicePacker(namespace string) mfd.Packer {
	return func(namespaces mfd.Namespaces) (interface{}, error) {
		return NewServiceTemplatePackage(namespace, namespaces, g.options), nil
	}
}

func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, false)
	if err != nil {
		return err
	}

	if len(g.options.Namespaces) != 0 {
		var filteredNameSpaces mfd.Namespaces
		for _, ns := range g.options.Namespaces {
			if p := project.Namespace(ns); p != nil {
				filteredNameSpaces = append(filteredNameSpaces, p)
			}
		}
		project.Namespaces = filteredNameSpaces
	}

	// generating model & converters for all namespaces
	// TODO separate params file?
	output := path.Join(g.options.Output, "model.go")
	if _, err := mfd.PackAndSave(project.Namespaces, output, modelTemplate, g.ModelPacker(), true); err != nil {
		return xerrors.Errorf("generate vt model error: %w", err)
	}

	output = path.Join(g.options.Output, "converter.go")
	if _, err := mfd.PackAndSave(project.Namespaces, output, converterTemplate, g.ModelPacker(), true); err != nil {
		return xerrors.Errorf("generate vt converter error: %w", err)
	}

	for _, ns := range project.Namespaces {
		// generating each namespace in separate file
		// getting file name without dots
		output := path.Join(g.options.Output, mfd.GoFileName(ns.Name)+".go")
		if _, err := mfd.PackAndSave(project.Namespaces, output, serviceTemplate, g.ServicePacker(ns.Name), true); err != nil {
			return xerrors.Errorf("generate service %s error: %w", ns, err)
		}
	}

	if err := RenderAndPrint(project.Namespaces, serverTemplate, g.ModelPacker()); err != nil {
		return xerrors.Errorf("generate vt server error: %w", err)
	}

	return nil
}

func RenderAndPrint(namespaces mfd.Namespaces, tmpl string, packer mfd.Packer) error {
	parsed, err := template.New("base").Parse(tmpl)
	if err != nil {
		return xerrors.Errorf("parsing template error: %w", err)
	}

	pack, err := packer(namespaces)
	if err != nil {
		return xerrors.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return xerrors.Errorf("processing model template error: %w", err)
	}

	fmt.Print(buffer.String())
	return nil
}
