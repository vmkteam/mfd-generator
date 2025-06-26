package xml

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	genna "github.com/dizzyfool/genna/lib"
	"github.com/dizzyfool/genna/model"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	mfdFlag         = "mfd"
	nssFlag         = "namespaces"
	printFlag       = "print"
	goPGVerFlag     = "gopgver"
	verboseFlag     = "verbose"
	quietFlag       = "quiet"
	customTypesFlag = "custom-types"

	quietAll = "all"
	quietNew = "new"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("xml", "Create or update project base with namespaces and entities", New())
}

// Generator represents mfd generator
type Generator struct {
	options Options
	verbose bool

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

	flags.StringSliceP(base.Tables, "t", []string{}, "table names for model generation separated by comma\nuse 'schema_name.*' to generate model for every table in model")

	flags.StringP(nssFlag, "n", "", "use this parameter to set table & namespace in format \"users=users,projects;shop=orders,prices\"")

	flags.IntP(goPGVerFlag, "g", 9, "go-pg version")

	flags.StringP(quietFlag, "q", "", "quiet mode. ignored when --namespaces (-n) flag is set. possible values:\n- all - will use namespace entity mapping from mfd, entities not present in mfd file will be ignored\n- new - generator will prompt namespace for entities not present in mfd file")

	flags.StringSlice(customTypesFlag, []string{}, "set custom types separated by comma\nformat: <postgresql_type>:<go_import>.<go_type>\nexamples: uuid:github.com/google/uuid.UUID,point:src/model.Point,bytea:string\n")

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

	// print namespaces
	if g.printNamespaces, err = flags.GetBool(printFlag); err != nil {
		return err
	}

	// table to process
	if g.options.Tables, err = flags.GetStringSlice(base.Tables); err != nil {
		return
	}

	// go-pg version
	if g.options.GoPgVer, err = flags.GetInt(goPGVerFlag); err != nil {
		return
	}

	if g.options.GoPgVer < mfd.GoPG8 || g.options.GoPgVer > mfd.GoPG10 {
		return fmt.Errorf("unsupported go-pg version: %d", g.options.GoPgVer)
	}

	// custom types
	var customTypesStrings []string
	if customTypesStrings, err = flags.GetStringSlice(customTypesFlag); err != nil {
		return err
	}

	if g.options.CustomTypes, err = model.ParseCustomTypes(customTypesStrings); err != nil {
		return err
	}

	// table to process
	if g.options.Quiet, err = flags.GetString(quietFlag); err != nil {
		return
	}

	if g.options.Quiet != quietAll && g.options.Quiet != quietNew && g.options.Quiet != "" {
		return fmt.Errorf(`unsupported quiet mode: %s, use "all" or "new"`, g.options.Quiet)
	}

	// preset packages
	nss, err := flags.GetString(nssFlag)
	if err != nil {
		return
	}

	if nss != "" {
		g.options.Packages = parseNamespacesFlag(nss)
	}

	if len(g.options.Tables) == 0 {
		// fill tables from namespaces if not set
		if len(g.options.Packages) != 0 {
			for table := range g.options.Packages {
				g.options.Tables = append(g.options.Tables, table)
			}
		} else {
			g.options.Tables = []string{"public.*"}
		}
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
	project, err := mfd.LoadProject(g.options.Output, true, g.options.GoPgVer)
	if err != nil {
		return err
	}

	// printing namespaces string
	if g.printNamespaces {
		//nolint:forbidigo
		fmt.Print(PrintNamespaces(project))
		return nil
	}

	addedCustomTypes := project.AddCustomTypes(g.options.CustomTypes)

	// reading tables from db
	entities, err := genna.Read(g.options.Tables, false, false, project.GoPGVer, project.CustomTypeMapping())
	if err != nil {
		return fmt.Errorf("read database, err=%w", err)
	}

	set := mfd.NewSet()
	// filling set
	for _, namespace := range project.Namespaces {
		set.Append(namespace.Name)
	}

	for _, entity := range entities {
		exiting := project.EntityByTable(entity.PGFullName)
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
			switch g.options.Quiet {
			case quietAll:
				if exiting != nil {
					namespace = exiting.Namespace
					break // case
				}
				continue // loop
			case quietNew:
				if exiting != nil {
					namespace = exiting.Namespace
					break // case
				}
				fallthrough // to default
			default:
				// asking namespace from prompt
				if namespace, err = g.PromptNS(entity.PGFullName, set.Elements()); err != nil {
					// may happen only in ctrl+c
					return fmt.Errorf("prompt namespace, err=%w", err)
				}
				// if user choose to skip
				if namespace == "skip" {
					continue // loop
				}
			}
		}

		// adding to set
		set.Prepend(namespace)

		// adding to project
		project.AddEntity(namespace, PackEntity(namespace, entity, exiting, addedCustomTypes))
	}

	// suggesting searches && fk links
	project.SuggestArrayLinks()
	project.UpdateLinks()

	// validate names
	if err := project.ValidateNames(); err != nil {
		return err
	}

	if err := project.IsConsistent(); err != nil {
		return fmt.Errorf("%w. fk table should be either be in project or selected for generatation", err)
	}

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
	formats := make([]string, len(project.Namespaces))
	for i := range project.Namespaces {
		format := make([]string, len(project.Namespaces[i].Entities))
		for j := range project.Namespaces[i].Entities {
			format[j] = project.Namespaces[i].Entities[j].Table
		}

		formats[i] = fmt.Sprintf("%s:%s", project.Namespaces[i].Name, strings.Join(format, ","))
	}

	return strings.Join(formats, ";")
}
