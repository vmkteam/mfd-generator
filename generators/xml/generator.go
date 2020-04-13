package xml

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/dizzyfool/genna/lib"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	mfdFlag     = "mfd"
	nssFlag     = "namespaces"
	printFlag   = "print"
	verboseFlag = "verbose"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml", "Create or update project base with namespaces and entities", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
	verbose bool
	base    base.Generator

	printNamespaces bool
}

// New creates generator
func New() *Generator {
	return &Generator{}
}

// AddFlags adds flags to command
func (g *Generator) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.BoolP(verboseFlag, "v", false, "print sql queries")

	flags.StringP(base.Conn, "c", "", "connection string to postgres database, e.g. postgres://usr:pwd@localhost:5432/db")
	if err := command.MarkFlagRequired(base.Conn); err != nil {
		panic(err)
	}

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(base.Tables, "t", []string{"public.*"}, "table names for model generation separated by comma\nuse 'schema_name.*' to generate model for every table in model")

	flags.StringP(nssFlag, "n", "", "use this parameter to set table & namespace in format \"users=users,projects;shop=orders,prices\"")

	flags.BoolP(printFlag, "p", false, "print namespace - tables association")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) (err error) {
	flags := command.Flags()

	if g.verbose, err = flags.GetBool(verboseFlag); err != nil {
		return
	}

	// connection to db
	if g.options.URL, err = flags.GetString(base.Conn); err != nil {
		return
	}

	// filepath to project model
	if g.options.Output, err = flags.GetString(mfdFlag); err != nil {
		return
	}

	// tables to process
	if g.options.Tables, err = flags.GetStringSlice(base.Tables); err != nil {
		return
	}

	if g.printNamespaces, err = flags.GetBool(printFlag); err != nil {
		return err
	}

	// preset packages
	nss, err := flags.GetString(nssFlag)
	if err != nil {
		return
	}

	if nss != "" {
		g.options.Packages = parseNamespacesFlag(nss)
	}

	return
}

func parseNamespacesFlag(v string) map[string]string {
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

// Generate runs generator
func (g *Generator) Generate() (err error) {
	var logger *log.Logger
	if g.verbose {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	genna := genna.New(g.options.URL, logger)

	// loading project from file
	project, err := mfd.LoadProject(g.options.Output, true)
	if err != nil {
		return err
	}

	// printing namespaces string
	if g.printNamespaces {
		fmt.Print(PrintNamespaces(project))
		return nil
	}

	// reading tables from db
	entities, err := genna.Read(g.options.Tables, false, false, project.GoPGVer)
	if err != nil {
		return fmt.Errorf("read database error: %w", err)
	}

	set := mfd.NewSet()
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

		// adding to project
		project.AddEntity(namespace, PackEntity(namespace, entity, exiting))
	}

	// suggesting searches
	project.SuggestArrayLinks()

	// saving mfd file
	if err = mfd.SaveMFD(g.options.Output, project); err != nil {
		return err
	}

	return mfd.SaveProjectXML(g.options.Output, project)
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
		AddLabel: "add new",
	}

	_, result, err = prompts.Run()

	return result, err
}

func PrintNamespaces(project *mfd.Project) string {
	var formats []string

	for _, namespace := range project.Namespaces {
		var format []string
		for _, entity := range namespace.Entities {
			format = append(format, entity.Table)
		}
		formats = append(formats, fmt.Sprintf("%s:%s", namespace, strings.Join(format, ",")))
	}

	return strings.Join(formats, ";")
}
