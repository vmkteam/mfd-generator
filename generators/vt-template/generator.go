package vttmpl

import (
	"bytes"
	"fmt"
	"html/template"
	"path"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	nsFlag       = "namespaces"
	entitiesFlag = "entities"

	routesTemplateFlag = "routes-tmpl"
	listTemplateFlag   = "list-tmpl"
	filterTemplateFlag = "filter-tmpl"
	formTemplateFlag   = "form-tmpl"
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

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces to generate. separate by comma\n")
	flags.StringSliceP(entitiesFlag, "e", []string{}, "entities to generate, must be in vt.xml file. separate by comma")

	flags.String(routesTemplateFlag, "", "path to routes custom template")
	flags.String(listTemplateFlag, "", "path to list custom template")
	flags.String(filterTemplateFlag, "", "path to filter custom template")
	flags.String(formTemplateFlag, "", "path to form custom template\n")
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

	if g.options.Entities, err = flags.GetStringSlice(entitiesFlag); err != nil {
		return err
	}

	if g.options.RoutesTemplatePath, err = flags.GetString(routesTemplateFlag); err != nil {
		return err
	}
	if g.options.ListTemplatePath, err = flags.GetString(listTemplateFlag); err != nil {
		return err
	}
	if g.options.FiltersTemplatePath, err = flags.GetString(filterTemplateFlag); err != nil {
		return err
	}
	if g.options.FiltersTemplatePath, err = flags.GetString(formTemplateFlag); err != nil {
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

	// validate names
	if err := project.ValidateNames(); err != nil {
		return err
	}

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	// loading templates
	routesTemplate, err := mfd.LoadTemplate(g.options.RoutesTemplatePath, routesDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load routes template error: %w", err)
	}

	listTemplate, err := mfd.LoadTemplate(g.options.ListTemplatePath, listDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load list template error: %w", err)
	}

	filterTemplate, err := mfd.LoadTemplate(g.options.FiltersTemplatePath, filterDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load filter template error: %w", err)
	}

	formTemplate, err := mfd.LoadTemplate(g.options.ListTemplatePath, formDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load form template error: %w", err)
	}

	// generating routes for all namespaces
	output := path.Join(g.options.Output, "src/pages/Entity/routes.ts")
	if _, err := SaveRoutes(project.VTNamespaces, output, routesTemplate); err != nil {
		return fmt.Errorf("generate routes error: %w", err)
	}

	translations, err := mfd.LoadTranslations(g.options.MFDPath, project.Languages)
	if err != nil {
		return fmt.Errorf("read translation error: %w", err)
	}

	for _, namespace := range g.options.Namespaces {
		ns := project.VTNamespace(namespace)
		if ns == nil {
			return fmt.Errorf("namespace %s not found in project", namespace)
		}

		entityNames := ns.VTEntityNames()
		if len(g.options.Entities) != 0 {
			entityNames = g.options.Entities
		}

		for _, name := range entityNames {
			entity := ns.VTEntity(name)
			if entity == nil {
				return fmt.Errorf("vt entity %s not found in project", name)
			}

			// skip if read only or none
			if entity.Mode == mfd.ModeReadOnly || entity.Mode == mfd.ModeNone {
				continue
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "List.vue")
			if err := SaveEntity(*entity, output, listTemplate); err != nil {
				return fmt.Errorf("generate entity %s list error: %w", entity.Name, err)
			}

			output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "components/ListFilters.vue")
			if err := SaveEntity(*entity, output, filterTemplate); err != nil {
				return fmt.Errorf("generate entity %s filters  error: %w", entity.Name, err)
			}

			// do not generate form on
			if entity.Mode != mfd.ModeReadOnlyWithTemplates {
				output = path.Join(g.options.Output, "src/pages/Entity", entity.Name, "Form.vue")
				if err := SaveEntity(*entity, output, formTemplate); err != nil {
					return fmt.Errorf("generate entity %s form  error: %w", entity.Name, err)
				}
			}

			// saving translations
			for lang, translation := range translations {
				output := path.Join(g.options.Output, "src/pages/Entity", entity.Name, lang+".json")
				if tre := translation.Entity(ns.Name, entity.Name); tre != nil {
					if err := mfd.MarshalJSONToFile(output, tre); err != nil {
						return fmt.Errorf("save translation lang %s error: %w", lang, err)
					}
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
func SaveRoutes(namespaces []*mfd.VTNamespace, output, tmpl string) (bool, error) {
	parsed, err := template.New("base").Funcs(mfd.TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, fmt.Errorf("parsing template error: %w", err)
	}

	pack, err := PackRoutesNamespace(namespaces)
	if err != nil {
		return false, fmt.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return false, fmt.Errorf("processing model template error: %w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}
