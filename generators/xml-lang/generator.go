package xmllang

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	langsFlag    = "langs"
	nssFlag      = "namespaces"
	entitiesFlag = "entities"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml-lang", "Create lang xml from mfd", New())
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

	flags.StringSliceP(langsFlag, "l", []string{}, "languages to generate, use two letters code, eg. ru,en,de. separate by comma")

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

	if g.options.Languages, err = flags.GetStringSlice(langsFlag); err != nil {
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
		g.options.Namespaces = project.VTNamespaceNames()
	}

	mfd.AddCustomTranslations(project.Dictionary)
	langs := mergeLangs(project.Languages, g.options.Languages)

	translations, err := mfd.LoadTranslations(g.options.MFDPath, langs)
	if err != nil {
		return fmt.Errorf("read translations, err=%w", err)
	}

	for lang, translation := range translations {
		for _, namespace := range g.options.Namespaces {
			ns := project.VTNamespace(namespace)
			if ns == nil {
				return fmt.Errorf("namespace %s not found", namespace)
			}

			entities := g.options.Entities
			if len(entities) == 0 {
				entities = ns.VTEntityNames()
			}

			translation = Translate(ns, translation, entities, lang)

			if err := mfd.SaveTranslation(translation, g.options.MFDPath, lang); err != nil {
				return fmt.Errorf("save translation lang %s, err=%w", lang, err)
			}
		}
	}

	project.Languages = langs
	return mfd.SaveMFD(g.options.MFDPath, project)
}

func mergeLangs(project, input []string) []string {
	set := mfd.NewSet()

	for _, lang := range project {
		set.Append(lang)
	}
	for _, lang := range input {
		set.Append(lang)
	}

	return set.Elements()
}
