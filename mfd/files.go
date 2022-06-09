package mfd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dizzyfool/genna/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func raw(str string) template.HTML {
	return template.HTML(str)
}

func title(str string) template.HTML {
	c := cases.Title(language.Und, cases.NoLower)
	return template.HTML(c.String(str))
}

var TemplateFunctions = template.FuncMap{
	"raw":     raw,
	"ToLower": strings.ToLower,
	"title":   title,
}

type Packer func(*Namespace) (interface{}, error)

// Load MFD Project from File
func LoadProject(filename string, create bool, goPGVer int) (*Project, error) {
	if _, err := os.Stat(filename); create && os.IsNotExist(err) {
		return NewProject(filepath.Base(filename), goPGVer), nil
	}

	project := &Project{}
	if err := UnmarshalFile(filename, project); err != nil {
		return nil, fmt.Errorf("read project error: %w", err)
	}

	project.Namespaces = []*Namespace{}
	project.VTNamespaces = []*VTNamespace{}

	dir := filepath.Dir(filename)
	for _, pf := range project.NamespaceNames {
		ns, err := LoadNamespace(path.Join(dir, pf+".xml"))
		if err != nil {
			return nil, fmt.Errorf("read namespace error: %w", err)
		}

		vtns, err := LoadVTNamespace(path.Join(dir, pf+".vt.xml"))
		if err != nil {
			return nil, fmt.Errorf("read vt vtns error: %w", err)
		}

		if len(ns.Entities) != 0 {
			project.Namespaces = append(project.Namespaces, ns)
		}

		if len(vtns.Entities) != 0 {
			project.VTNamespaces = append(project.VTNamespaces, vtns)
		}
	}

	// backward compatibility
	if len(project.Languages) == 0 {
		project.Languages = []string{EnLang}
	}
	if project.GoPGVer == 0 {
		project.GoPGVer = goPGVer
	}

	project.UpdateLinks()

	return project, project.IsConsistent()
}

func LoadNamespace(filename string) (*Namespace, error) {
	namespace := &Namespace{}
	if err := UnmarshalFile(filename, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

func LoadVTNamespace(filename string) (*VTNamespace, error) {
	namespace := &VTNamespace{
		Name: strings.TrimSuffix(path.Base(filename), ".vt.xml"),
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return namespace, nil
	}

	if err := UnmarshalFile(filename, namespace); err != nil {
		return nil, err
	}

	// backward comp
	if namespace.Name == "" {
		namespace.Name = strings.TrimSuffix(path.Base(filename), ".vt.xml")
	}

	return namespace, nil
}

func UnmarshalFile(filename string, v interface{}) (err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	if err := xml.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func SaveMFD(filename string, p *Project) error {
	if err := MarshalToFile(filename, p); err != nil {
		return fmt.Errorf("save project error: %w", err)
	}

	return nil
}

func SaveProjectXML(filename string, p *Project) error {
	for _, namespace := range p.Namespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".xml")
		if err := MarshalToFile(file, namespace); err != nil {
			return fmt.Errorf("save namespace %s error: %w", namespace.Name, err)
		}
	}

	return nil
}

func SaveProjectVT(filename string, p *Project) error {
	for _, namespace := range p.VTNamespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".vt.xml")
		if err := MarshalToFile(file, namespace); err != nil {
			return fmt.Errorf("save namespace vt entites %s error: %w", namespace.Name, err)
		}
	}

	return nil
}

func MarshalToFile(filename string, v interface{}) error {
	bytes, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	if _, err := Save(bytes, filename); err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	return nil
}

func FormatAndSave(data interface{}, output, tmpl string, format bool) (bool, error) {
	parsed, err := template.New("base").Funcs(TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, fmt.Errorf("parsing template error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", data); err != nil {
		return false, fmt.Errorf("processing model template error: %w", err)
	}

	if format {
		return util.FmtAndSave(buffer.Bytes(), output)
	}

	return Save(buffer.Bytes(), output)
}

func GoFileName(namespace string) string {
	if parts := strings.SplitN(namespace, ".", 2); len(parts) >= 2 {
		return strings.ToLower(parts[1])
	}

	return strings.ToLower(namespace)
}

func Save(content []byte, filename string) (bool, error) {
	file, err := util.File(filename)
	if err != nil {
		return false, fmt.Errorf("open model file error: %w", err)
	}

	if _, err := file.Write(content); err != nil {
		return false, fmt.Errorf("writing content to file error: %w", err)
	}

	return true, nil
}

func LoadTranslations(project string, languages []string) (map[string]Translation, error) {
	translations := map[string]Translation{}

	for _, lang := range languages {
		translation := Translation{
			Language: lang,
		}

		filename := path.Join(filepath.Dir(project), lang+".xml")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			translations[lang] = translation
			continue
		}

		if err := UnmarshalFile(filename, &translation); err != nil {
			return nil, err
		}

		translations[lang] = translation
	}

	return translations, nil
}

func SaveTranslation(translation Translation, project, language string) error {
	filename := path.Join(filepath.Dir(project), language+".xml")
	return MarshalToFile(filename, translation)
}

func MarshalJSONToFile(filename string, v interface{}) error {
	bytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	if _, err := Save(bytes, filename); err != nil {
		return err
	}
	return nil
}

func LoadTemplate(path, def string) (string, error) {
	if strings.Trim(path, " ") == "" {
		return def, nil
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
