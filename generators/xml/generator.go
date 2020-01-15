package xml

import (
	"fmt"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/dizzyfool/genna/lib"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

const packages = "pkgs"

// CreateCommand creates generator command
func CreateCommand(logger *zap.Logger) *cobra.Command {
	return base.CreateCommand("xml", "Create xml from database", New(logger))
}

// Generator represents mfd generator
type Generator struct {
	logger  *zap.Logger
	options Options
	base    base.Generator
}

// New creates generator
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

	flags.StringP(base.Conn, "c", "", "connection string to your postgres database")
	if err := command.MarkFlagRequired(base.Conn); err != nil {
		panic(err)
	}

	flags.StringP(base.Output, "o", "", "output mfd file name")
	if err := command.MarkFlagRequired(base.Output); err != nil {
		panic(err)
	}

	flags.StringSliceP(base.Tables, "t", []string{"public.*"}, "table names for model generation separated by comma\nuse 'schema_name.*' to generate model for every table in model")

	flags.StringP(packages, "p", "", "use this parameter to set table & namespace in format \"users=users,projects;shop=orders,prices\"")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) (err error) {
	flags := command.Flags()

	// connection to db
	if g.options.URL, err = flags.GetString(base.Conn); err != nil {
		return
	}

	// filepath to project model
	if g.options.Output, err = flags.GetString(base.Output); err != nil {
		return
	}

	// tables to process
	if g.options.Tables, err = flags.GetStringSlice(base.Tables); err != nil {
		return
	}

	// preset packages
	pkgs, err := flags.GetString(packages)
	if err != nil {
		return
	}

	if pkgs != "" {
		g.options.Packages = parsePackagesParam(pkgs)
	}

	return
}

func parsePackagesParam(v string) map[string]string {
	// processing format
	// namespace1:table1,table2;namespace2:table3

	mp := map[string]string{}

	namespaces := strings.Split(v, ";")
	for _, namespace := range namespaces {
		if parts := strings.Split(namespace, ":"); len(parts) == 2 {
			name := parts[0]
			tables := strings.Split(parts[1], ",")
			for _, table := range tables {
				mp[table] = name
			}
		}
	}

	if len(mp) == 0 {
		return nil
	}

	return mp
}

// Generate runs whole generation process
func (g *Generator) Generate() (err error) {
	genna := genna.New(g.options.URL, g.logger)

	// reading tables from db
	entities, err := genna.Read(g.options.Tables, false, false)
	if err != nil {
		return xerrors.Errorf("read database error: %w", err)
	}

	// loading project from file
	project, err := mfd.LoadProject(g.options.Output, true)
	if err != nil {
		return err
	}

	set := mfd.NewSet()
	nse := map[string]mfd.Entities{}

	for _, entity := range entities {
		exiting := project.Entity(entity.GoName)
		if exiting != nil {
			set.Prepend(exiting.Namespace)
		}

		var namespace string

		if g.options.Packages != nil {
			// getting namespace from preset
			var ok bool
			if namespace, ok = g.options.Packages[entity.PGFullName]; !ok {
				continue
			}
		} else {
			// asking namespace from prompt
			if namespace, err = g.PromptNS(entity.PGFullName, set.Elements()); err != nil {
				// may happen only in ctrl+c
				return nil
			}
			// if user choose to skip
			if namespace == "skip" {
				continue
			}
		}

		// adding to set
		set.Prepend(namespace)

		// adding to map
		nse[namespace] = append(nse[namespace], PackEntity(namespace, entity))
	}

	// adding entities to project
	for namespace, entities := range nse {
		for i, entity := range entities {
			nse[namespace][i] = project.Namespace(namespace).AddEntity(entity)
		}
	}

	// updating links
	project.UpdateLinks()
	project.SuggestArrayLinks()

	// adding vt entities to project
	for namespace, entities := range nse {
		for i, entity := range entities {
			nse[namespace][i].VTEntity = project.Namespace(namespace).AddVTEntity(PackVTEntity(entity))
		}
	}

	// adding vt templates to project
	for namespace, entities := range nse {
		for i, entity := range entities {
			nse[namespace][i].VTEntity.AddTmpl(PackTemplate(entity, entity.VTEntity))
		}
	}

	// saving translation
	translations, err := mfd.LoadTranslations(g.options.Output, []string{mfd.RuLang, mfd.EnLang})
	for lang, translation := range translations {
		if err != nil {
			return xerrors.Errorf("read translation lang %s error: %w", lang, err)
		}
		translation.Merge(Translate(project, lang))
		if err := mfd.SaveTranslation(translation, g.options.Output, lang); err != nil {
			return xerrors.Errorf("save translation lang %s error: %w", lang, err)
		}
	}

	// saving files
	return mfd.SaveProject(g.options.Output, project)
}

// PromptNS prompting namespace in console
func (g *Generator) PromptNS(table string, namespaces []string) (result string, err error) {
	// skipping statuses
	if table == "statuses" {
		return "skip", nil
	}

	prompts := promptui.SelectWithAdd{
		Label:    fmt.Sprintf("Choose namespace for table %s", table),
		Items:    append(namespaces, "skip"),
		AddLabel: "or add new",
	}

	_, result, err = prompts.Run()

	return result, err
}
