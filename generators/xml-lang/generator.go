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
	return base.CreateCommand("xml-lang", "Create lang xml from mfd", New())
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

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(langsFlag, "l", []string{}, "languages to generate, use two letters code, eg. ru,en,de. separate by comma")
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

	langs := mergeLangs(project.Languages, g.options.Languages)

	translations, err := mfd.LoadTranslations(g.options.MFDPath, langs)
	for lang, translation := range translations {
		if err != nil {
			return fmt.Errorf("read translation lang %s error: %w", lang, err)
		}
		Translate(project, &translation, lang)

		if err := mfd.SaveTranslation(translation, g.options.MFDPath, lang); err != nil {
			return fmt.Errorf("save translation lang %s error: %w", lang, err)
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
