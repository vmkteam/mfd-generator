package dbtest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/dizzyfool/genna/util"
	"github.com/spf13/cobra"
)

const (
	mfdFlag   = "mfd"
	pkgFlag   = "package"
	dbPkgFlag = "dbPkg"
	nssFlag   = "namespaces"
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

	flags.StringP(base.Output, "o", "", "output dir path")
	if err := command.MarkFlagRequired(base.Output); err != nil {
		panic(err)
	}

	flags.StringP(mfdFlag, "m", "", "mfd file path")
	if err := command.MarkFlagRequired(mfdFlag); err != nil {
		panic(err)
	}

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used")

	flags.StringP(dbPkgFlag, "x", "", "package containing db files got with model generator")
	if err := command.MarkFlagRequired(dbPkgFlag); err != nil {
		panic(err)
	}

	flags.StringSliceP(nssFlag, "n", []string{}, "namespaces to generate. separate by comma\n")
}

// ReadFlags reads basic flags from command
func (g *Generator) ReadFlags(command *cobra.Command) (err error) {
	flags := command.Flags()

	if g.options.Output, err = flags.GetString(base.Output); err != nil {
		return err
	}

	if g.options.MFDPath, err = flags.GetString(mfdFlag); err != nil {
		return err
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

	g.options.Def()

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

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	// Prepare template
	tmpl, err := mfd.LoadTemplate("", funcTemplate)
	if err != nil {
		return fmt.Errorf("load func template, err=%w", err)
	}

	parsedTmpl, err := template.New("base").Funcs(mfd.TemplateFunctions).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing func template, err=%w", err)
	}

	for _, namespace := range g.options.Namespaces {
		// generating each func in separate files by namespace
		if ns := project.Namespace(namespace); ns != nil {
			// Generate test helpers
			err = g.generateFuncsByNS(ns, parsedTmpl)
			if err != nil {
				return fmt.Errorf("failed to generate test helpers: %w", err)
			}
		}
	}

	return nil
}

func (g *Generator) SaveSetupFile() (bool, error) {
	output := path.Join(g.options.Output, "test.go")
	if fileExists(output) {
		return false, nil
	}

	// Generate base file with conn initialization and helpers
	tmpl, err := mfd.LoadTemplate("", baseFileTemplate)
	if err != nil {
		return false, fmt.Errorf("load model template, err=%w", err)
	}

	parsed, err := template.New("base").Funcs(mfd.TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, fmt.Errorf("parsing template, err=%w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", PackFuncRenderData(g.options)); err != nil {
		return false, fmt.Errorf("processing model template, err=%w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

func (g *Generator) CreateFuncFile(nsName string) (bool, error) {
	// Getting file name without dots
	output := filepath.Join(g.options.Output, mfd.GoFileName(nsName)+".go")
	// Generate base file with conn initialization and helpers
	tmpl, err := mfd.LoadTemplate("", funcFileTemplate)
	if err != nil {
		return false, fmt.Errorf("load func file template, err=%w", err)
	}

	parsed, err := template.New("funcFile").Funcs(mfd.TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, fmt.Errorf("parsing func file template, err=%w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "funcFile", PackFuncRenderData(g.options)); err != nil {
		return false, fmt.Errorf("processing func file template, err=%w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

// generateFuncsByNS generates the test helper functions
func (g *Generator) generateFuncsByNS(ns *mfd.Namespace, tmpl *template.Template) error {
	// Getting file name without dots
	output := filepath.Join(g.options.Output, mfd.GoFileName(ns.Name)+".go")

	// Walk for each namespace and check if they have already had file and extract function names from them
	// Note: consider that the func names are distinct across all namespaces (because they have the same pkg)
	existingFunctions := make(map[string]struct{})
	if !fileExists(output) { // If file doesn't exist, create it
		if _, err := g.CreateFuncFile(ns.Name); err != nil {
			return fmt.Errorf("create file for functions, ns=%s, err=%w", ns.Name, err)
		}
	}

	// Parse namespace file and fetch existing func names
	existing, err := g.parseExistingFunctions(output)
	if err != nil {
		return fmt.Errorf("parse existing functions, ns=%s, err=%w", ns.Name, err)
	}

	// Set existing functions to our map
	for i := range existing {
		existingFunctions[existing[i]] = struct{}{}
	}

	buffer := new(bytes.Buffer)
	nsData := PackNamespace(ns, g.options)
	// Render funcs for each entity
	for _, entity := range nsData.Entities {
		if _, ok := existingFunctions[entity.Name]; ok {
			continue // Skip if the function already exists
		}

		// Render func
		if err := tmpl.ExecuteTemplate(buffer, "base", entity); err != nil {
			return fmt.Errorf("processing func template, err=%w", err)
		}
	}

	if buffer.Len() == 0 {
		return nil
	}

	content, err := os.ReadFile(output)
	if err != nil {
		return fmt.Errorf("read file=%s, err=%w", output, err)
	}

	content = append(content, buffer.Bytes()...)
	// Write to file
	if _, err := util.FmtAndSave(content, output); err != nil {
		return fmt.Errorf("fmt and write to file=%s, err=%w", output, err)
	}

	return nil
}

// parseExistingFunctions parses existing helper functions to avoid overwriting,
// returns slice of func names
func (g *Generator) parseExistingFunctions(output string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, output, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file, path=%s, err=%w", output, err)
	}

	var res []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			res = append(res, fn.Name.Name)
		}
		return true
	})

	return res, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
