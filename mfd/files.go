package mfd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
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

// LoadProject Loads MFD Project from File
func LoadProject(filename string, create bool, goPGVer int) (*Project, error) {
	if _, err := os.Stat(filename); create && os.IsNotExist(err) {
		return NewProject(filepath.Base(filename), goPGVer), nil
	}

	project := &Project{}
	if err := UnmarshalFile(filename, project); err != nil {
		return nil, fmt.Errorf("read project, err=%w", err)
	}

	project.Namespaces = []*Namespace{}
	project.VTNamespaces = []*VTNamespace{}

	dir := filepath.Dir(filename)
	for _, pf := range project.NamespaceNames {
		ns, err := LoadNamespace(path.Join(dir, pf+".xml"))
		//todo: maybe create xmls if not exists
		if err != nil {
			return nil, fmt.Errorf("read namespace, err=%w", err)
		}

		vtns, err := LoadVTNamespace(path.Join(dir, pf+".vt.xml"))
		if err != nil {
			return nil, fmt.Errorf("read vt vtns, err=%w", err)
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
		return fmt.Errorf("read file, err=%w", err)
	}

	if err := xml.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("unmarshal, err=%w", err)
	}

	return nil
}

func SaveMFD(filename string, p *Project) error {
	if err := MarshalToFile(filename, p); err != nil {
		return fmt.Errorf("save project, err=%w", err)
	}

	return nil
}

func SaveProjectXML(filename string, p *Project) error {
	for _, namespace := range p.Namespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".xml")
		if err := MarshalToFile(file, namespace); err != nil {
			return fmt.Errorf("save namespace %s, err=%w", namespace.Name, err)
		}
	}

	return nil
}

func SaveProjectVT(filename string, p *Project) error {
	for _, namespace := range p.VTNamespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".vt.xml")
		if err := MarshalToFile(file, namespace); err != nil {
			return fmt.Errorf("save namespace vt entites %s, err=%w", namespace.Name, err)
		}
	}

	return nil
}

func MarshalToFile(filename string, v interface{}) error {
	b, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal data, err=%w", err)
	}

	// need for json searching rules
	b = bytes.ReplaceAll(b, []byte("-&gt;"), []byte("->"))

	// append line break at end of file, if not exists
	if !bytes.HasSuffix(b, []byte("\n")) {
		b = append(b, []byte("\n")...)
	}

	if _, err = Save(b, filename); err != nil {
		return fmt.Errorf("write file, err=%w", err)
	}

	return nil
}

func FormatAndSave(data interface{}, output, tmpl string, format bool) (bool, error) {
	buffer, err := renderTemplate(data, tmpl)
	if err != nil {
		return false, fmt.Errorf("render template, err=%w", err)
	}

	if format {
		return util.FmtAndSave(buffer.Bytes(), output)
	}

	return Save(buffer.Bytes(), output)
}

// renderTemplate generates data based on a template and returns it in a buffer.
//
// Parameters:
//   - data: the data used to populate the template.
//   - tmpl: the string containing the template for data generation.
//
// Returns:
//   - bytes.Buffer: a buffer containing the generated data.
//   - error: an error if there was an issue parsing or executing the template.
//
// Example usage:
//
//		data := ... // data for the template
//		tmpl := "Template string with {{.}}"
//		buffer, err := renderTemplate(data, tmpl)
//		if err != nil {
//	    log.Fatal(err)
//		}
//		fmt.Println("Rendered template:", buffer.String())
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

// replaceFragmentInFile replaces fragments of text in a file with new data if they match a given regular expression.
//
// Parameters:
// - output: the path to the file where the replacement will take place.
// - findData: the string to search for within the fragment to be replaced.
// - newData: the string to replace the found fragment with.
// - pattern: a pointer to a compiled regular expression used to find the fragments.
//
// Returns:
//
// - bool: true if the replacement was successful, and false otherwise.
//
// - error: an error if there was an issue reading the file or replacing the fragments.
//
// Example usage:
//
//		output := "path/to/file.txt"
//		findData := "old text"
//		newData := "new text"
//		pattern := regexp.MustCompile(`pattern`)
//		success, err := replaceFragmentInFile(output, findData, newData, pattern)
//		if err != nil {
//	    log.Fatal(err)
//		}
//
// fmt.Println("Replacement successful:", success)
func replaceFragmentInFile(output, findData, newData string, pattern *regexp.Regexp) (bool, error) {
	content, err := os.ReadFile(output)
	if err != nil {
		return false, fmt.Errorf("read file err: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	ff, err := extractFragments(pattern, lines)
	if err != nil {
		return false, fmt.Errorf("extract fragments error: %w", err)
	}

	var replaced bool
	for _, fragment := range ff {
		s, end := fragment[0], fragment[1]
		extractedFragment := lines[s:end]
		for _, extline := range extractedFragment {
			if strings.Contains(extline, findData) {
				var resultLines []string
				resultLines = append(resultLines, lines[:s]...)
				resultLines = append(resultLines, strings.Split(newData, "\n")...)
				resultLines = append(resultLines, lines[end:]...)

				lines = resultLines
				replaced = true
				break
			}
		}
		if replaced {
			break
		}
	}

	if !replaced {
		lines = append(lines, strings.Split(newData, "\n")...)
	}

	newContent := strings.Join(lines, "\n")
	return util.FmtAndSave([]byte(newContent), output)
}

// extractFragments searches for fragments of text in the provided lines that match the given regular expression.
// It returns an array of tuples, each containing the start and end coordinates of the text fragments found.
//
// Parameters:
//
// - re: a pointer to a compiled regular expression used to find matches in the lines.
//
// - lines: an array of strings to search for matching text fragments.
//
// Returns:
//
// - [][2]int: an array of tuples, each containing the start and end indices of the found text fragments.
//
// - error: an error if no matching fragments are found.
//
// Example usage:
//
//		re := regexp.MustCompile(`pattern`)
//		lines := []string{"text1", "text2", "pattern", "text3", "}"}
//		fragments, err := extractFragments(re, lines)
//		if err != nil {
//	    log.Fatal(err)
//		}
//
// fmt.Println(fragments) // [[2, 5]]
func extractFragments(re *regexp.Regexp, lines []string) ([][2]int, error) {
	var (
		reFragments [][2]int
		start       = -1
	)

	for i, line := range lines {
		if re.MatchString(line) {
			if start != -1 {
				// added fragment in slice
				reFragments = append(reFragments, [2]int{start, i})
			}
			// first coordinate in fragment
			start = i
		}
	}

	// if the last fragment is not closed in the loop
	// think that this fragment is to the end of the lines
	if start != -1 {
		reFragments = append(reFragments, [2]int{start, len(lines)})
	}

	if len(reFragments) == 0 {
		return nil, errors.New("no reFragments found with pattern")
	}

	var ff [][2]int

	// split big fragment
	for _, fragment := range reFragments {
		ll := lines[fragment[0]:fragment[1]]
		var subStart = fragment[0]

		for i, line := range ll {
			if line == "}" {
				ff = append(ff, [2]int{subStart, fragment[0] + i + 1})
				subStart = fragment[0] + i + 1
			}
		}
	}

	return ff, nil
}

// UpdateFile updates a file by replacing specific fragments with new data generated from a template.
//
// Parameters:
//   - data: the data used to populate the template.
//   - output: the path to the file where the replacement will take place.
//   - tmpl: the path to the template file.
//   - pattern: a pointer to a compiled regular expression used to find the fragments.
//
// Returns:
//   - bool: true if the update was successful, and false otherwise.
//   - error: an error if there was an issue generating the data, replacing the fragments, or at other stages.
//
// Example usage:
//
//		data := ... // data for the template
//		output := "path/to/file.txt"
//		tmpl := "path/to/template.tmpl"
//		pattern := regexp.MustCompile(`pattern`)
//		success, err := UpdateFile(data, output, tmpl, pattern)
//		if err != nil {
//	    log.Fatal(err)
//		}
//
// fmt.Println("Update successful:", success)
func UpdateFile(data interface{}, output, tmpl string, pattern *regexp.Regexp) (bool, error) {
	buffer, err := renderTemplate(data, tmpl)
	if err != nil {
		return false, fmt.Errorf("generating data error: %w", err)
	}

	// get []string from generate template
	lines := strings.Split(buffer.String(), "\n")

	// search fragment from our template
	fragments, err := extractFragments(pattern, lines)
	if err != nil {
		return false, fmt.Errorf("extract fragments err=%w", err)
	}

	for _, fragment := range fragments {
		var filePart []string
		var findRow string

		filePart = append(filePart, lines[fragment[0]:fragment[1]]...)
		for _, part := range filePart {
			if isStartedRow(part) {
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

// isStartedRow find row where start function or structure
func isStartedRow(row string) bool {
	if (strings.Contains(row, "func") || strings.Contains(row, "struct")) && !strings.Contains(row, "//") {
		return true
	}
	return false
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
		return false, fmt.Errorf("open model file, err=%w", err)
	}

	if _, err := file.Write(content); err != nil {
		return false, fmt.Errorf("writing content to file, err=%w", err)
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
		return fmt.Errorf("marshal data, err=%w", err)
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

func AddCustomTranslations(d *Dictionary) {
	if d != nil {
		for _, e := range d.Entries {
			presetsTranslations[RuLang][e.XMLName.Local] = e.Value
		}
	}
}
