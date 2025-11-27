package dbtest

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	pkgFlag      = "package"
	dbPkgFlag    = "db-pkg"
	nssFlag      = "namespaces"
	entitiesFlag = "entities"
	forceFlag    = "force"

	FuncPattern = `^func (\w+)`
)

var (
	funcRe = regexp.MustCompile(FuncPattern)
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("dbtest", "Create or update functions from xml for inserting testdata into tables", New())
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

	flags.StringSliceP(nssFlag, "n", []string{}, "namespaces to generate. Separate by comma\n")
	flags.StringSliceP(entitiesFlag, "e", []string{}, "entities to generate. Separate by comma\n")

	flags.BoolP(forceFlag, "f", false, "force generate if functions already exist. Deletes old and generates new functions")
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

	if g.options.Entities, err = flags.GetStringSlice(entitiesFlag); err != nil {
		return err
	}

	if g.options.Force, err = flags.GetBool(forceFlag); err != nil {
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
	g.options.CustomTypes = project.CustomTypes

	// validate names
	if err := project.ValidateNames(); err != nil {
		return err
	}

	if _, err := g.SaveSetupFile(); err != nil {
		return fmt.Errorf("generate setup file, err=%w", err)
	}

	// If there is nothing provided in the namespaces and entities, we should generate all entities by all namespaces
	if len(g.options.Namespaces) == 0 && len(g.options.Entities) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	// Set namespace names as keys. It can be all if they're not provided
	for _, ns := range g.options.Namespaces {
		g.options.EntitiesByNamespace[ns] = nil
	}

	// Check if we don't have enough namespaces by provided entities.
	// For instance, somebody runs:
	//  dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd' -e news,region -n portal
	// where news belongs portal but region does not. We have to put regions' namespace "geo"
	for _, entity := range g.options.Entities {
		nsName := project.Entity(entity).Namespace
		if !slices.Contains(g.options.Namespaces, nsName) {
			g.options.Namespaces = append(g.options.Namespaces, nsName)
		}
		entities, ok := g.options.EntitiesByNamespace[nsName]
		if !ok || len(entities) == 0 {
			g.options.EntitiesByNamespace[nsName] = []string{entity}
			continue
		}

		if slices.Contains(entities, entity) {
			continue
		}

		g.options.EntitiesByNamespace[nsName] = append(g.options.EntitiesByNamespace[nsName], entity)
	}

	for _, namespace := range g.options.Namespaces {
		// Walk through each namespace and check if they have already had file and extract function names from them
		// Note: consider that the func names are distinct across all namespaces (because they have the same pkg)
		if ns := project.Namespace(namespace); ns != nil {
			// Generate test helpers
			err = g.generateFuncsByNS(ns)
			if err != nil {
				return fmt.Errorf("failed to generate test helpers: %w", err)
			}
		}
	}

	return nil
}

func (g *Generator) SaveSetupFile() (bool, error) {
	output := path.Join(g.options.Output, "test.go")
	isForce := g.options.Force && (len(g.options.Namespaces) == 0 || len(g.options.Entities) == 0)
	if !isForce && fileExists(output) {
		return false, nil
	}

	buffer := new(bytes.Buffer)
	if err := mfd.RenderText(buffer, baseFileTemplate, PackFuncRenderData(g.options)); err != nil {
		return false, fmt.Errorf("processing setup file template, err=%w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

func (g *Generator) CreateFuncFile(ns NamespaceData) (bool, error) {
	// Getting file name without dots
	output := filepath.Join(g.options.Output, mfd.GoFileName(ns.Name)+".go")

	buffer := new(bytes.Buffer)
	if err := mfd.Render(buffer, funcFileTemplate, ns); err != nil {
		return false, fmt.Errorf("processing func file template, err=%w", err)
	}

	return mfd.Save(buffer.Bytes(), output)
}

// generateFuncsByNS generates the test helper functions
func (g *Generator) generateFuncsByNS(ns *mfd.Namespace) error {
	// Getting file name without dots
	output := filepath.Join(g.options.Output, mfd.GoFileName(ns.Name)+".go")
	nsData := PackNamespace(ns, g.options)
	entitiesToGenerate := g.options.EntitiesByNamespace[ns.Name]
	checkEntities := len(entitiesToGenerate) > 0 && !nsData.HasAllOfProvidedEntities(entitiesToGenerate)

	if !fileExists(output) || (len(entitiesToGenerate) == 0 && g.options.Force) {
		if _, err := g.CreateFuncFile(nsData); err != nil {
			return fmt.Errorf("create file for functions, ns=%s, err=%w", ns.Name, err)
		}
	}

	entities := make(map[string]struct{}, len(g.options.Entities))
	for i := range g.options.Entities {
		entities[g.options.Entities[i]] = struct{}{}
	}

	// Render funcs for each entity
	for _, entity := range nsData.Entities {
		if _, ok := entities[entity.VarName]; !ok && checkEntities {
			continue // Skip if entities are provided and it is not one of them
		}

		// Render opFunc type struct
		// Make a regexp with entity name to prevent removing OpFunc types despite the entity
		typeOpFuncRe := regexp.MustCompile(fmt.Sprintf(`^type %[1]sOpFunc func\(t \*testing\.T, dbo orm\.DB, in \*%[2]s\.%[1]s\) Cleaner`, entity.Name, entity.DBPackageAlias))
		if err := g.replaceTargetFromFile(OpFuncType{}, typeOpFuncRe, entity, output, "", ""); err != nil {
			return fmt.Errorf("replace the main func, entity=%s, err=%w", entity.Name, err)
		}

		// Render the main func
		if err := g.replaceTargetFromFile(MainFunc{}, funcRe, entity, output, "{", "}"); err != nil {
			return fmt.Errorf("replace the main func, entity=%s, err=%w", entity.Name, err)
		}

		// Render WithRelations opFunc
		if err := g.replaceTargetFromFile(OpFuncWithRelations{}, funcRe, entity, output, "{", "}"); err != nil {
			return fmt.Errorf("replace the main func, entity=%s, err=%w", entity.Name, err)
		}

		// Render WithFake opFunc
		if err := g.replaceTargetFromFile(OpFuncWithFake{}, funcRe, entity, output, "{", "}"); err != nil {
			return fmt.Errorf("replace the main func, entity=%s, err=%w", entity.Name, err)
		}
	}

	return nil
}

// replaceTargetFromFile removes specified functions from a Go file
// The target could be func, struct or type OpFunc
func (g *Generator) replaceTargetFromFile(b FuncLayoutRenderer, targetReExpression *regexp.Regexp, entity EntityData, filePath, openingToken, closeningToken string) error {
	buf := new(bytes.Buffer)
	if err := b.Render(buf, entity); err != nil {
		return fmt.Errorf("render the main func, entity=%s, err=%w", entity.Name, err)
	}

	// If it rendered nothing, the main condition in the template doesn't work, just skip
	if buf.Len() == 0 {
		return nil
	}

	if _, err := mfd.UpdateFile(buf, filePath, openingToken, closeningToken, targetReExpression, g.options.Force); err != nil {
		return fmt.Errorf("update file, err=%w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
