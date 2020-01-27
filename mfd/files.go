package mfd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dizzyfool/genna/util"
	"golang.org/x/xerrors"
)

func raw(str string) template.HTML {
	return template.HTML(str)
}

var TemplateFunctions = template.FuncMap{
	"raw":     raw,
	"ToLower": strings.ToLower,
}

type Packer func(Namespaces) (interface{}, error)

// Load MFD Project from File
func LoadProject(filename string, create bool) (*Project, error) {
	if _, err := os.Stat(filename); create && os.IsNotExist(err) {
		return NewProject(filepath.Base(filename)), nil
	}

	project := &Project{}
	if err := UnmarshalFile(filename, project); err != nil {
		return nil, xerrors.Errorf("read project error: %w", err)
	}

	project.Namespaces = Namespaces{}
	project.Filename = filename

	dir := filepath.Dir(filename)
	for _, pf := range project.NamespaceNames {
		ns, err := LoadNamespace(path.Join(dir, pf+".xml"))
		if err != nil {
			return nil, xerrors.Errorf("read namespace error: %w", err)
		}

		entities, err := LoadVTEntities(path.Join(dir, pf+".vt.xml"))
		if err != nil {
			return nil, xerrors.Errorf("read vt entities error: %w", err)
		}

		for _, e := range entities.Entities {
			ns.AddVTEntity(e)
		}

		project.Namespaces = append(project.Namespaces, ns)
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

func LoadVTEntities(filename string) (*VTNamespace, error) {
	vtEntities := &VTNamespace{}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return vtEntities, nil
	}

	if err := UnmarshalFile(filename, vtEntities); err != nil {
		return nil, err
	}

	return vtEntities, nil
}

func UnmarshalFile(filename string, v interface{}) (err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return xerrors.Errorf("read file error: %w", err)
	}

	if err := xml.Unmarshal(bytes, v); err != nil {
		return xerrors.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func SaveMFD(filename string, p *Project) error {
	if err := MarshalToFile(filename, p); err != nil {
		return xerrors.Errorf("save project error: %w", err)
	}

	return nil
}

func SaveProjectXML(filename string, p *Project) error {
	for _, namespace := range p.Namespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".xml")
		if err := MarshalToFile(file, namespace); err != nil {
			return xerrors.Errorf("save namespace %s error: %w", namespace.Name, err)
		}
	}

	return nil
}

func SaveProjectVT(filename string, p *Project) error {
	for _, namespace := range p.Namespaces {
		file := path.Join(filepath.Dir(filename), namespace.Name+".vt.xml")
		if err := MarshalToFile(file, NewVTNamespace(namespace.VTEntities())); err != nil {
			return xerrors.Errorf("save namespace vt entites %s error: %w", namespace.Name, err)
		}
	}

	return nil
}

func MarshalToFile(filename string, v interface{}) error {
	bytes, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return xerrors.Errorf("marshal data error: %w", err)
	}

	if _, err := Save(bytes, filename); err != nil {
		return xerrors.Errorf("write file error: %w", err)
	}

	return nil
}

func PackAndSave(namespaces Namespaces, output, tmpl string, packer Packer, format bool) (bool, error) {
	parsed, err := template.New("base").Funcs(TemplateFunctions).Parse(tmpl)
	if err != nil {
		return false, xerrors.Errorf("parsing template error: %w", err)
	}

	pack, err := packer(namespaces)
	if err != nil {
		return false, xerrors.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return false, xerrors.Errorf("processing model template error: %w", err)
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
		return false, xerrors.Errorf("open model file error: %w", err)
	}

	if _, err := file.Write(content); err != nil {
		return false, xerrors.Errorf("writing content to file error: %w", err)
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

func MarshalJSONToFile(filename string, v TranslationEntity) error {
	bytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return xerrors.Errorf("marshal data error: %w", err)
	}

	if _, err := Save(bytes, filename); err != nil {
		return err
	}
	return nil
}
