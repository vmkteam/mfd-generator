package mfd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
	"notLast": func(index int, length int) bool {
		return index+1 != length
	},
	"isLast": func(index int, length int) bool {
		return index+1 == length
	},
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
		//todo: maybe create xmls if not exists
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
	if bytes, err = os.ReadFile(filename); err != nil {
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
	b, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	// need for json searching rules
	b = bytes.ReplaceAll(b, []byte("-&gt;"), []byte("->"))

	// append line break at end of file, if not exists
	if !bytes.HasSuffix(b, []byte("\n")) {
		b = append(b, []byte("\n")...)
	}

	if _, err = Save(b, filename); err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	return nil
}

func FormatAndSave(data interface{}, output, tmpl string, format bool) (bool, error) {
	buffer, err := renderTemplate(data, tmpl)
	if err != nil {
		return false, fmt.Errorf("generating data error: %w", err)
	}

	if format {
		return util.FmtAndSave(buffer.Bytes(), output)
	}

	return Save(buffer.Bytes(), output)
}

func renderTemplate(data interface{}, tmpl string) (bytes.Buffer, error) {
	var buffer bytes.Buffer
	parsed, err := template.New("base").Funcs(TemplateFunctions).Parse(tmpl)
	if err != nil {
		return buffer, fmt.Errorf("parsing template error: %w", err)
	}

	if err := parsed.ExecuteTemplate(&buffer, "base", data); err != nil {
		return buffer, fmt.Errorf("processing model template error: %w", err)
	}
	return buffer, nil
}

func replaceFragmentInFile(output, findData, newData, pattern string) (bool, error) {
	content, err := os.ReadFile(output)
	if err != nil {
		return false, fmt.Errorf("read file err: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	ff, err := extractFragments(pattern, lines)

	if err != nil {
		return false, fmt.Errorf("extract fragments error: %w", err)
	}

	var newlines []string
	for _, fragment := range ff {
		s, end := fragment[0], fragment[1]
		extractedFragment := lines[s:end]
		for _, extline := range extractedFragment {
			if strings.Contains(extline, findData) {
				newlines = append(newlines, lines[:s]...)
				newlines = append(newlines, strings.Split(newData, "\n")...)
				newlines = append(newlines, lines[end:]...)
				break
			}
		}
	}

	if len(newlines) == 0 {
		newlines = append(lines, strings.Split(newData, "\n")...)
	}

	newContent := strings.Join(newlines, "\n")

	err = os.WriteFile(output, []byte(newContent), 0644)
	if err != nil {
		return false, fmt.Errorf("err write in file: %w", err)
	}

	return true, nil
}

func extractFragments(pattern string, lines []string) ([][]int, error) {
	var (
		reFragments [][]int
		ff          [][]int
		start       = -1
	)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("regexp error: %w", err)
	}
	for i, line := range lines {
		if re.MatchString(line) {
			if start != -1 {
				reFragments = append(reFragments, []int{start, i})
			}
			start = i
		}
	}

	if start != -1 {
		reFragments = append(reFragments, []int{start, len(lines)})
	}

	if len(reFragments) == 0 {
		return nil, fmt.Errorf("no reFragments found with pattern: %s", pattern)
	}

	// split big fragment
	for _, fragment := range reFragments {
		ll := lines[fragment[0]:fragment[1]]
		var subStart = fragment[0]

		for i, line := range ll {
			if line == "}" {
				ff = append(ff, []int{subStart, fragment[0] + i + 1})
				subStart = fragment[0] + i + 1
			}
		}
	}

	return ff, nil
}

func UpdateFile(data interface{}, output, tmpl, pattern string) (bool, error) {
	buffer, err := renderTemplate(data, tmpl)
	if err != nil {
		return false, fmt.Errorf("generating data error: %w", err)
	}

	lines := strings.Split(buffer.String(), "\n")

	fragments, err := extractFragments(pattern, lines)

	for _, fragment := range fragments {
		var filePart []string
		var findRow string

		filePart = append(filePart, lines[fragment[0]:fragment[1]]...)
		for _, part := range filePart {
			if strings.Contains(part, "func") || strings.Contains(part, "struct {") {
				findRow = part
				break
			}
		}
		if _, err := replaceFragmentInFile(output, strings.TrimSuffix(findRow, " "), strings.Join(filePart, "\n"), pattern); err != nil {
			return false, fmt.Errorf("replace fragment error: %w", err)
		}
	}

	return true, nil
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

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
