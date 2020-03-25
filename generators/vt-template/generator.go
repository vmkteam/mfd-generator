package vttmpl

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag         = "mfd"
	nsFlag          = "ns"
	listTmplFlag    = "list-tmpl"
	filtersTmplFlag = "filter-tmpl"
	formTmplFlag    = "form-tmpl"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("template", "Create vt template from xml", New())
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

// Generate runs generator
func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, false)
	if err != nil {
		return err
	}

	if len(g.options.Namespaces) != 0 {
		var filteredNameSpaces []*mfd.Namespace
		for _, ns := range g.options.Namespaces {
			if p := project.Namespace(ns); p != nil {
				filteredNameSpaces = append(filteredNameSpaces, p)
			}
		}
		project.Namespaces = filteredNameSpaces
	}

	// generating routes for all namespaces
	output := path.Join(g.options.Output, "src/pages/Entity/routes.ts")
	if _, err := SaveRoutes(project.VTNamespaces, output); err != nil {
		return fmt.Errorf("generate routes error: %w", err)
	}

	// loading templates
	listTmpl := listTemplate
	if g.options.ListTemplate != "" {
		listTmpl, err = loadTemplate(g.options.ListTemplate)
		if err != nil {
			return fmt.Errorf("load list template error: %w", err)
		}
	}

	filtersTmpl := filterTemplate
	if g.options.FiltersTemplate != "" {
		filtersTmpl, err = loadTemplate(g.options.FiltersTemplate)
		if err != nil {
			return fmt.Errorf("load filters template error: %w", err)
		}
	}

	formTmpl := formTemplate
	if g.options.FormTemplate != "" {
		listTmpl, err = loadTemplate(g.options.FormTemplate)
		if err != nil {
			return fmt.Errorf("load form template error: %w", err)
		}
	}

	translations, err := mfd.LoadTranslations(g.options.MFDPath, project.Languages)
	if err != nil {
		return fmt.Errorf("read translation error: %w", err)
	}

	for _, namespace := range project.VTNamespaces {
		for _, entity := range namespace.Entities {
			if entity.NoTemplates {
				continue
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "List.vue")
			if err := SaveEntity(*entity, output, listTmpl); err != nil {
				return fmt.Errorf("generate entity %s list error: %w", entity.Name, err)
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "components/ListFilters.vue")
			if err := SaveEntity(*entity, output, filtersTmpl); err != nil {
				return fmt.Errorf("generate entity %s filters  error: %w", entity.Name, err)
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "Form.vue")
			if err := SaveEntity(*entity, output, formTmpl); err != nil {
				return fmt.Errorf("generate entity %s form  error: %w", entity.Name, err)
			}

			// saving translations
			for lang, translation := range translations {
				output := path.Join(g.options.Output, "src/pages/Entity", entity.Name, lang+".json")
				if err := mfd.MarshalJSONToFile(output, translation.Entity(namespace.Name, entity.Name)); err != nil {
					return fmt.Errorf("save translation lang %s error: %w", lang, err)
				}
			}
		}
	}

	return mfd.SaveMFD(g.options.MFDPath, project)
}

// SaveEntity saves vt entity to template with special delims
func SaveEntity(entity mfd.VTEntity, output, tmpl string) error {
	parsed, err := template.New("base").
		Delims("[[", "]]").
		Funcs(mfd.TemplateFunctions).
		Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template error: %w", err)
	}

	packed := PackEntity(entity)

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", packed); err != nil {
		return fmt.Errorf("processing model template error: %w", err)
	}

	_, err = mfd.Save(buffer.Bytes(), output)
	return err
}

// SaveRoutes saves all vt namespaces to routes file
func SaveRoutes(namespaces []*mfd.VTNamespace, output string) (bool, error) {
	parsed, err := template.New("base").Funcs(mfd.TemplateFunctions).Parse(routesTemplate)
	if err != nil {
		return false, fmt.Errorf("parsing template error: %w", err)
	}

	pack, err := NewTemplatePackage(namespaces)
	if err != nil {
		return false, fmt.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return false, fmt.Errorf("processing model template error: %w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

func loadTemplate(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
