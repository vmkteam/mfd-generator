package vt

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	pkgFlag      = "package"
	modelPkgFlag = "model"
	nsFlag       = "namespaces"
	embedLogFlag = "embedlog-pkg"

	modelTemplateFlag     = "model-tmpl"
	converterTemplateFlag = "converter-tmpl"
	serviceTemplateFlag   = "service-tmpl"
	serverTemplateFlag    = "server-tmpl"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("vt", "Create vt from xml", New())
}

// Generator represents mfd vt generator
type Generator struct {
	options Options
}

// New creates vt generator
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

	flags.StringP(modelPkgFlag, "x", "", "package containing model files got with model generator")
	if err := command.MarkFlagRequired(modelPkgFlag); err != nil {
		panic(err)
	}

	flags.StringP(embedLogFlag, "l", "", "package containing embedlog. if not set - it wil be detected from model path")

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used")

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces to generate. separate by comma\n")

	flags.String(modelTemplateFlag, "", "path to model custom template")
	flags.String(converterTemplateFlag, "", "path to converter custom template")
	flags.String(serviceTemplateFlag, "", "path to service custom template")
	flags.String(serverTemplateFlag, "", "path to server custom template\n")
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

	if g.options.EmbedLogPackage, err = flags.GetString(embedLogFlag); err != nil {
		return err
	}

	if g.options.EmbedLogPackage == "" {
		g.options.EmbedLogPackage = strings.TrimSuffix(g.options.ModelPackage, path.Base(g.options.ModelPackage)) + "embedlog"
	}

	if g.options.Package, err = flags.GetString(pkgFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nsFlag); err != nil {
		return err
	}

	if g.options.Package == "" {
		g.options.Package = path.Base(g.options.Output)
	}

	if g.options.ModelTemplatePath, err = flags.GetString(modelTemplateFlag); err != nil {
		return err
	}
	if g.options.ConverterTemplatePath, err = flags.GetString(converterTemplateFlag); err != nil {
		return err
	}
	if g.options.ServiceTemplatePath, err = flags.GetString(serviceTemplateFlag); err != nil {
		return err
	}
	if g.options.ServerTemplatePath, err = flags.GetString(serverTemplateFlag); err != nil {
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

	modelTemplate, err := mfd.LoadTemplate(g.options.ModelTemplatePath, modelDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load model template error: %w", err)
	}

	converterTemplate, err := mfd.LoadTemplate(g.options.ConverterTemplatePath, converterDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load converter template error: %w", err)
	}

	serviceTemplate, err := mfd.LoadTemplate(g.options.ServiceTemplatePath, serviceDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load service template error: %w", err)
	}

	serverTemplate, err := mfd.LoadTemplate(g.options.ServerTemplatePath, serverDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load server template error: %w", err)
	}

	for _, namespace := range g.options.Namespaces {
		ns := project.VTNamespace(namespace)
		if ns == nil {
			return fmt.Errorf("namespace %s not found in project", namespace)
		}

		// generating each namespace in separate file
		baseName := mfd.GoFileName(ns.Name)

		modelData, err := PackNamespace(ns, g.options)
		if err != nil {
			return fmt.Errorf("generate vt model error: %w", err)
		}

		// generate model file
		output := path.Join(g.options.Output, fmt.Sprintf("%s_model.go", baseName))
		if _, err := mfd.FormatAndSave(modelData, output, modelTemplate, true); err != nil {
			return fmt.Errorf("generate vt model error: %w", err)
		}

		// generate converter file
		output = path.Join(g.options.Output, fmt.Sprintf("%s_converter.go", baseName))
		if _, err := mfd.FormatAndSave(modelData, output, converterTemplate, true); err != nil {
			return fmt.Errorf("generate vt converter error: %w", err)
		}

		// generate service file
		output = path.Join(g.options.Output, fmt.Sprintf("%s.go", baseName))
		serviceData := PackServiceNamespace(ns, g.options)
		if _, err := mfd.FormatAndSave(serviceData, output, serviceTemplate, true); err != nil {
			return fmt.Errorf("generate service %s error: %w", namespace, err)
		}
	}

	// printing zenrpc server code
	if err := PrintServer(project.VTNamespaces, serverTemplate, g.options); err != nil {
		return fmt.Errorf("generate vt server error: %w", err)
	}

	return nil
}

func PrintServer(namespaces []*mfd.VTNamespace, tmpl string, options Options) error {
	parsed, err := template.New("base").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template error: %w", err)
	}

	pack, err := PackServerNamespaces(namespaces, options)
	if err != nil {
		return fmt.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return fmt.Errorf("processing model template error: %w", err)
	}

	fmt.Print(buffer.String())
	return nil
}
