package vt

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"regexp"
	"slices"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/generators/base"
	"github.com/spf13/cobra"
)

const (
	mfdFlag      = "mfd"
	pkgFlag      = "package"
	modelPkgFlag = "model"
	nsFlag       = "namespaces"
	embedLogFlag = "embedlog-pkg"
	entityFlag   = "entities"

	modelTemplateFlag     = "model-tmpl"
	converterTemplateFlag = "converter-tmpl"
	serviceTemplateFlag   = "service-tmpl"
	serverTemplateFlag    = "server-tmpl"

	defaultLoggerPkg = "github.com/vmkteam/embedlog"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("vt", "Create vt from xml", New())
}

// Generator represents mfd vt generator
type Generator struct {
	options Options
}

// New creates vt generator
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

	flags.StringP(modelPkgFlag, "x", "", "package containing model files got with model generator")
	if err := command.MarkFlagRequired(modelPkgFlag); err != nil {
		panic(err)
	}

	flags.StringP(embedLogFlag, "l", "", "package containing embedlog. if not set - it wil be detected from model path")

	flags.StringP(pkgFlag, "p", "", "package name that will be used in golang files. if not set - last element of output path will be used")

	flags.StringSliceP(nsFlag, "n", []string{}, "namespaces to generate. separate by comma\n")
	flags.StringSliceP(entityFlag, "e", []string{}, "entities to generate. separate by comma\n")

	flags.String(modelTemplateFlag, "", "path to model custom template")
	flags.String(converterTemplateFlag, "", "path to converter custom template")
	flags.String(serviceTemplateFlag, "", "path to service custom template")
	flags.String(serverTemplateFlag, "", "path to server custom template\n")
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

	if g.options.ModelPackage, err = flags.GetString(modelPkgFlag); err != nil {
		return err
	}

	if g.options.EmbedLogPackage, err = flags.GetString(embedLogFlag); err != nil {
		return err
	}

	if g.options.EmbedLogPackage == "" {
		g.options.EmbedLogPackage = defaultLoggerPkg
	}

	if g.options.Package, err = flags.GetString(pkgFlag); err != nil {
		return err
	}

	if g.options.Namespaces, err = flags.GetStringSlice(nsFlag); err != nil {
		return err
	}

	if g.options.Entities, err = flags.GetStringSlice(entityFlag); err != nil {
		return err
	}

	if g.options.Package == "" {
		g.options.Package = path.Base(g.options.Output)
	}

	if g.options.ModelTemplatePath, err = flags.GetString(modelTemplateFlag); err != nil {
		return err
	}
	if g.options.ConverterTemplatePath, err = flags.GetString(converterTemplateFlag); err != nil {
		return err
	}
	if g.options.ServiceTemplatePath, err = flags.GetString(serviceTemplateFlag); err != nil {
		return err
	}
	if g.options.ServerTemplatePath, err = flags.GetString(serverTemplateFlag); err != nil {
		return err
	}

	g.options.Def()

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

	g.options.GoPGVer = project.GoPGVer
	g.options.CustomTypes = project.CustomTypes

	if len(g.options.Namespaces) == 0 {
		g.options.Namespaces = project.NamespaceNames
	}

	modelTemplate, err := mfd.LoadTemplate(g.options.ModelTemplatePath, modelDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load model template, err=%w", err)
	}

	converterTemplate, err := mfd.LoadTemplate(g.options.ConverterTemplatePath, converterDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load converter template, err=%w", err)
	}

	serviceTemplate, err := mfd.LoadTemplate(g.options.ServiceTemplatePath, serviceDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load service template, err=%w", err)
	}

	serverTemplate, err := mfd.LoadTemplate(g.options.ServerTemplatePath, serverDefaultTemplate)
	if err != nil {
		return fmt.Errorf("load server template, err=%w", err)
	}

	if len(g.options.Entities) > 0 {
		return g.PartialUpdate(project, modelTemplate, converterTemplate, serviceTemplate)
	}

	for _, namespace := range g.options.Namespaces {
		ns := project.VTNamespace(namespace)
		if ns == nil {
			return fmt.Errorf("namespace %s not found in project", namespace)
		}

		// generating each namespace in separate file
		baseName := mfd.GoFileName(ns.Name)

		modelData, err := PackNamespace(ns, g.options)
		if err != nil {
			return fmt.Errorf("generate vt model, err=%w", err)
		}

		// generate model file
		output := path.Join(g.options.Output, fmt.Sprintf("%s_model.go", baseName))
		if _, err := mfd.FormatAndSave(modelData, output, modelTemplate, true); err != nil {
			return fmt.Errorf("generate vt model, err=%w", err)
		}

		// generate converter file
		output = path.Join(g.options.Output, fmt.Sprintf("%s_converter.go", baseName))
		if _, err := mfd.FormatAndSave(modelData, output, converterTemplate, true); err != nil {
			return fmt.Errorf("generate vt converter, err=%w", err)
		}

		// generate service file
		output = path.Join(g.options.Output, fmt.Sprintf("%s.go", baseName))
		serviceData, err := PackServiceNamespace(ns, g.options)
		if err != nil {
			return fmt.Errorf("pack service namespace, err=%w", err)
		}
		if _, err := mfd.FormatAndSave(serviceData, output, serviceTemplate, true); err != nil {
			return fmt.Errorf("generate service %s, err=%w", namespace, err)
		}
	}

	// printing zenrpc server code
	if err := PrintServer(project.VTNamespaces, serverTemplate, g.options); err != nil {
		return fmt.Errorf("generate vt server, err=%w", err)
	}

	return nil
}

func PrintServer(namespaces []*mfd.VTNamespace, tmpl string, options Options) error {
	parsed, err := template.New("base").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template, err=%w", err)
	}

	pack, err := PackServerNamespaces(namespaces, options)
	if err != nil {
		return fmt.Errorf("packing data, err=%w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return fmt.Errorf("processing model template, err=%w", err)
	}

	//nolint:forbidigo
	fmt.Print(buffer.String())
	return nil
}

// PartialUpdate updates parts of the project by generating model, converter, and service files based on specified templates.
//
// Parameters:
//   - project: a pointer to the mfd.Project to be updated.
//   - modelTemplate: the path to the model template file.
//   - converterTemplate: the path to the converter template file.
//   - serviceTemplate: the path to the service template file.
//
// Returns:
//   - error: an error if there was an issue updating the files or at other stages.
//
// Example usage:
//
//		project := ... // initialize the project
//		modelTemplate := "path/to/model_template.tmpl"
//		converterTemplate := "path/to/converter_template.tmpl"
//		serviceTemplate := "path/to/service_template.tmpl"
//		err := g.PartialUpdate(project, modelTemplate, converterTemplate, serviceTemplate)
//		if err != nil {
//	    log.Fatal(err)
//		}
func (g *Generator) PartialUpdate(project *mfd.Project, modelTemplate, converterTemplate, serviceTemplate string) error {
	for _, namespace := range g.options.Namespaces {
		ns := project.VTNamespace(namespace)
		if ns == nil {
			return fmt.Errorf("namespace %s not found in project", namespace)
		}

		serviceData, err := PackServiceNamespace(ns, g.options)
		if err != nil {
			return fmt.Errorf("pack service namespace, err=%w", err)
		}

		ee, err := g.TargetServiceEntityData(serviceData)
		if err != nil {
			return err
		}

		modelData, err := PackNamespace(ns, g.options)
		if err != nil {
			return fmt.Errorf("generate vt model,err=%w", err)
		}

		eem, err := g.TargetModelEntityData(modelData)
		if err != nil {
			return err
		}

		baseName := mfd.GoFileName(ns.Name)

		// precompile regexp
		funcRe := regexp.MustCompile(FuncPattern)
		structRe := regexp.MustCompile(StructPattern)

		for _, e := range ee {
			// generate service file
			output := path.Join(g.options.Output, fmt.Sprintf("%s.go", baseName))
			serviceData.Entities = []ServiceEntityData{e}
			buf := new(bytes.Buffer)
			if err := mfd.Render(buf, serviceTemplate, serviceData); err != nil {
				return fmt.Errorf("render service, err=%w", err)
			}
			if _, err := mfd.UpdateFile(buf, output, "{", "}", structRe, true); err != nil {
				return fmt.Errorf("update file, service=%s, err=%w", namespace, err)
			}
		}

		for _, entity := range eem {
			// generate model file
			output := path.Join(g.options.Output, fmt.Sprintf("%s_model.go", baseName))
			modelData.Entities = []EntityData{entity}
			modelBuf := new(bytes.Buffer)
			if err := mfd.Render(modelBuf, modelTemplate, modelData); err != nil {
				return fmt.Errorf("render model, model=%s, err=%w", entity.Name, err)
			}
			if _, err := mfd.UpdateFile(modelBuf, output, "{", "}", structRe, true); err != nil {
				return fmt.Errorf("update model file, entity=%s, err=%w", entity.Name, err)
			}
			// generate converter file
			converterBuf := new(bytes.Buffer)
			if err := mfd.Render(converterBuf, converterTemplate, modelData); err != nil {
				return fmt.Errorf("render converer, model=%s, err=%w", entity.Name, err)
			}
			output = path.Join(g.options.Output, fmt.Sprintf("%s_converter.go", baseName))
			if _, err := mfd.UpdateFile(converterBuf, output, "{", "}", funcRe, true); err != nil {
				return fmt.Errorf("update converter file, entity=%s, err=%w", entity.Name, err)
			}
		}
	}

	return nil
}

// TargetServiceEntityData filters and returns the model entity data that are located in the specified namespace.
//
// Parameters:
//   - s: a ServiceNamespaceData object containing information about the namespace and its entities.
//
// Returns:
//   - [ ] ServiceEntityData: an array of entity data that are located in the specified namespace.
//   - error: an error if no entities were found or if some entities were not found in the specified namespace.
func (g *Generator) TargetServiceEntityData(s ServiceNamespaceData) ([]ServiceEntityData, error) {
	var ee []ServiceEntityData
	le := len(g.options.Entities)
	for _, e := range s.Entities {
		if slices.Contains(g.options.Entities, e.VarName) {
			ee = append(ee, e)
		}
	}
	if len(ee) == 0 && le > 0 {
		return nil, fmt.Errorf("namespace %s contains entities but no entity found", s.Name)
	}
	if len(ee) < le {
		return nil, fmt.Errorf("namespace %s contains entities but not all located in thes namespace", s.Name)
	}

	return ee, nil
}

// TargetModelEntityData filters and returns the model entity data that are located in the specified namespace.
//
// Parameters:
//   - s: a NamespaceData object containing information about the namespace and its entities.
//
// Returns:
//   - [ ] EntityData: an array of entity data that are located in the specified namespace.
//   - error: an error if no entities were found or if some entities were not found in the specified namespace.
func (g *Generator) TargetModelEntityData(s NamespaceData) ([]EntityData, error) {
	var ee []EntityData
	le := len(g.options.Entities)
	for _, e := range s.Entities {
		if slices.Contains(g.options.Entities, e.VarName) {
			ee = append(ee, e)
		}
	}
	if len(ee) == 0 && le > 0 {
		return nil, fmt.Errorf("namespace %s contains entities but no entity found", s.ModelPackage)
	}
	if len(ee) < le {
		return nil, fmt.Errorf("namespace %s contains entities but not all located in thes namespace", s.ModelPackage)
	}

	return ee, nil
}
