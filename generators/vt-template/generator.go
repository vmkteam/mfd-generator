package vttmpl

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

const (
	mfdFlag         = "mfd"
	nsFlag          = "ns"
	listTmplFlag    = "list-tmpl"
	filtersTmplFlag = "filter-tmpl"
	formTmplFlag    = "form-tmpl"
)

// CreateCommand creates generator command
func CreateCommand(logger *zap.Logger) *cobra.Command {
	return base.CreateCommand("template", "Create vt template from xml", New(logger))
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

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces")

	flags.StringP(listTmplFlag, "l", "", "path to file with list template")
	flags.StringP(filtersTmplFlag, "f", "", "path to file with filters template")
	flags.StringP(formTmplFlag, "d", "", "path to file with form template")
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

	if g.options.ListTemplate, err = flags.GetString(listTmplFlag); err != nil {
		return err
	}

	if g.options.FiltersTemplate, err = flags.GetString(formTmplFlag); err != nil {
		return err
	}

	if g.options.FormTemplate, err = flags.GetString(formTmplFlag); err != nil {
		return err
	}

	return nil
}

// Packer returns packer function for compile entities into package
func (g *Generator) FactoryPacker() mfd.Packer {
	return func(namespaces mfd.Namespaces) (interface{}, error) {
		return NewTemplatePackage(namespaces, g.options)
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

	// generating routes for all namespaces
	output := path.Join(g.options.Output, "src/pages/Entity/routes.ts")
	if _, err := mfd.PackAndSave(project.Namespaces, output, routesTemplate, g.FactoryPacker(), false); err != nil {
		return xerrors.Errorf("generate vt model error: %w", err)
	}

	// loading templates
	listTmpl := listTemplate
	if g.options.ListTemplate != "" {
		listTmpl, err = loadTemplate(g.options.ListTemplate)
		if err != nil {
			return xerrors.Errorf("load list template error: %w", err)
		}
	}

	filtersTmpl := filterTemplate
	if g.options.FiltersTemplate != "" {
		filtersTmpl, err = loadTemplate(g.options.FiltersTemplate)
		if err != nil {
			return xerrors.Errorf("load filters template error: %w", err)
		}
	}

	formTmpl := formTemplate
	if g.options.FormTemplate != "" {
		listTmpl, err = loadTemplate(g.options.FormTemplate)
		if err != nil {
			return xerrors.Errorf("load form template error: %w", err)
		}
	}

	for _, namespace := range project.Namespaces {
		for _, entity := range namespace.Entities {
			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "List.vue")
			if err := SaveEntity(*entity, output, listTmpl); err != nil {
				return xerrors.Errorf("generate entity %s list error: %w", entity.Name, err)
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "components/ListFilters.vue")
			if err := SaveEntity(*entity, output, filtersTmpl); err != nil {
				return xerrors.Errorf("generate entity %s filters  error: %w", entity.Name, err)
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "Form.vue")
			if err := SaveEntity(*entity, output, formTmpl); err != nil {
				return xerrors.Errorf("generate entity %s form  error: %w", entity.Name, err)
			}
		}
	}

	// saving translation
	for _, lang := range []string{mfd.RuLang, mfd.EnLang} {
		translation, err := mfd.LoadTranslation(g.options.MFDPath, lang)
		if err != nil {
			return xerrors.Errorf("read translation lang %s error: %w", lang, err)
		}

		output := path.Join(g.options.Output, "src/locales", lang+".json")
		if err := mfd.MarshalJSONToFile(output, translation.JSON()); err != nil {
			return xerrors.Errorf("save translation lang %s error: %w", lang, err)
		}
	}

	return nil
}

func SaveEntity(entity mfd.Entity, output, tmpl string) error {
	parsed, err := template.New("base").
		Delims("[[", "]]").
		Funcs(mfd.TemplateFunctions).
		Parse(tmpl)
	if err != nil {
		return xerrors.Errorf("parsing template error: %w", err)
	}

	pack := NewVTTemplateEntity(entity)

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return xerrors.Errorf("processing model template error: %w", err)
	}

	if _, err := mfd.Save(buffer.Bytes(), output); err != nil {
		return err
	}

	return nil
}

func loadTemplate(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
