package dbtest

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag    = "mfd"
	pkgFlag    = "package"
	dbPkgFlag  = "dbPkg"
	nssFlag    = "namespaces"
	entityFlag = "entities"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("dbtest", "Create or update functions for insert testdata by namespaces and entities", New())
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

	flags.StringP(dbPkgFlag, "x", "", "package containing db files got with model generator")
	if err := command.MarkFlagRequired(dbPkgFlag); err != nil {
		panic(err)
	}

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used")

	flags.StringSliceP(nssFlag, "n", []string{}, "namespaces to generate. separate by comma\n")
	flags.StringSliceP(entityFlag, "e", []string{}, "entities to generate. separate by comma\n")
	flags.String(connTemplate, "", "path to search custom template")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) (err error) {
	flags := command.Flags()

	// filepath to project model
	if g.options.Output, err = flags.GetString(mfdFlag); err != nil {
		return
	}

	if g.options.GoPGVer < mfd.GoPG8 || g.options.GoPGVer > mfd.GoPG10 {
		return fmt.Errorf("unsupported go-pg version: %d", g.options.GoPGVer)
	}

	if g.options.Package, err = flags.GetString(pkgFlag); err != nil {
		return err
	}

	if g.options.Package == "" {
		g.options.Package = path.Base(g.options.Output)
	}

	if g.options.DBPackage, err = flags.GetString(dbPkgFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nssFlag); err != nil {
		return err
	}

	if g.options.Entities, err = flags.GetStringSlice(entityFlag); err != nil {
		return err
	}

	if g.options.ConnTemplatePath, err = flags.GetString(connTemplate); err != nil {
		return err
	}

	return
}

// Generate runs generator
func (g *Generator) Generate() (err error) {
	// loading project from file
	project, err := mfd.LoadProject(g.options.MFDPath, false, 0)
	if err != nil {
		return err
	}

	g.options.GoPGVer = project.GoPGVer
	g.options.ProjectName = strings.TrimSuffix(project.Name, filepath.Ext(project.Name)) // Trim extension

	// validate names
	if err := project.ValidateNames(); err != nil {
		return err
	}

	if _, err := g.SaveSetupFile(); err != nil {
		return fmt.Errorf("generate setup file, err=%w", err)
	}

	return nil
}

func (g *Generator) SaveSetupFile() (bool, error) {
	output := path.Join(g.options.Output, "test.go")
	if fileExists(output) {
		return false, nil
	}

	// Generate base file with conn initialization and helpers
	tmpl, err := mfd.LoadTemplate(g.options.ConnTemplatePath, connTemplate)
	if err != nil {
		return false, fmt.Errorf("load model template, err=%w", err)
	}

	parsed, err := template.New("base").Funcs(mfd.TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, fmt.Errorf("parsing template, err=%w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", PackNamespace(g.options)); err != nil {
		return false, fmt.Errorf("processing model template, err=%w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
