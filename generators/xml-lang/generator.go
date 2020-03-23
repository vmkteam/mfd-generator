package xmllang

import (
	"fmt"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag   = "mfd"
	langsFlag = "langs"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml-lang", "Create lang xml from database", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
	verbose bool
	base    base.Generator
}

// New creates generator
func New() *Generator {
	return &Generator{}
}

// AddFlags adds flags to command
func (g *Generator) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.StringP(mfdFlag, "m", "", "mfd file")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(langsFlag, "l", []string{mfd.RuLang, mfd.EnLang}, "namespaces")
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

	return nil
}

// Generate runs generator
func (g *Generator) Generate() error {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, true)
	if err != nil {
		return err
	}

	translations, err := mfd.LoadTranslations(g.options.MFDPath, g.options.Languages)
	for lang, translation := range translations {
		if err != nil {
			return fmt.Errorf("read translation lang %s error: %w", lang, err)
		}
		translation.Merge(Translate(project, lang))
		if err := mfd.SaveTranslation(translation, g.options.MFDPath, lang); err != nil {
			return fmt.Errorf("save translation lang %s error: %w", lang, err)
		}
	}

	return nil
}
