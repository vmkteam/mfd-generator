package model

import (
	"embed"
	"fmt"
	"os"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var content embed.FS

const (
	mfdFlag = "mfd"
	pkgFlag = "package"

	modelTemplateFlag    = "model-tmpl"
	validateTemplateFlag = "validate-tmpl"
	searchTemplateFlag   = "search-tmpl"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("model", "Create golang model from xml", New())
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

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used\n")

	flags.String(modelTemplateFlag, "", "path to model custom template")
	flags.String(searchTemplateFlag, "", "path to search custom template")
	flags.String(validateTemplateFlag, "", "path to validate custom template\n")
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

	if g.options.Package == "" {
		g.options.Package = path.Base(g.options.Output)
	}

	if g.options.ModelTemplatePath, err = flags.GetString(modelTemplateFlag); err != nil {
		return err
	}
	if g.options.SearchTemplatePath, err = flags.GetString(searchTemplateFlag); err != nil {
		return err
	}
	if g.options.ValidateTemplatePath, err = flags.GetString(validateTemplateFlag); err != nil {
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

	g.options.GoPGVer = project.GoPGVer
	g.options.CustomTypes = project.CustomTypes

	// validate names
	if err := project.ValidateNames(); err != nil {
		return err
	}

	// basic generator
	output := path.Join(g.options.Output, "model.go")
	modelData := PackNamespace(project.Namespaces, g.options)
	modelTemplate, err := mfd.LoadTemplate(g.options.ModelTemplatePath, modelDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load model template, err=%w", err)
	}

	if _, err := mfd.FormatAndSave(modelData, output, modelTemplate, true); err != nil {
		return fmt.Errorf("generate project model, err=%w", err)
	}

	// generating search
	output = path.Join(g.options.Output, "model_search.go")
	searchData, err := PackSearchNamespace(project.Namespaces, g.options)
	if err != nil {
		return fmt.Errorf("pack search namespace, err=%w", err)
	}
	searchTemplate, err := mfd.LoadTemplate(g.options.SearchTemplatePath, searchDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load model template, err=%w", err)
	}

	if _, err := mfd.FormatAndSave(searchData, output, searchTemplate, true); err != nil {
		return fmt.Errorf("generate project search, err=%w", err)
	}

	// generating validate
	output = path.Join(g.options.Output, "model_validate.go")
	validateDate := PackValidateNamespace(project.Namespaces, g.options)
	validateTemplate, err := mfd.LoadTemplate(g.options.ValidateTemplatePath, validateDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load model template, err=%w", err)
	}

	// generating base db files
	for _, file := range []string{"db.go", "filter.go", "filter_json.go", "options.go"} {
		p := path.Join(g.options.Output, file)

		// check file existence
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			continue
		}

		// generate file if not exists
		b, err := content.ReadFile(fmt.Sprintf("templates/%s.tmpl", file))
		if err != nil {
			return fmt.Errorf("read model template, err=%w", err)
		}

		if _, err = mfd.Save(b, p); err != nil {
			return fmt.Errorf("save model template, err=%w", err)
		}
	}

	if _, err := mfd.FormatAndSave(validateDate, output, validateTemplate, true); err != nil {
		return fmt.Errorf("generate project validate, err=%w", err)
	}

	// generating params
	output = path.Join(g.options.Output, "model_params.go")
	if _, err := GenerateParams(project.Namespaces, output, g.options); err != nil {
		return fmt.Errorf("generate project params, err=%w", err)
	}

	return nil
}
