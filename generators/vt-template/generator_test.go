package vttmpl

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"
	"github.com/vmkteam/mfd-generator/mfd"

	. "github.com/smartystreets/goconvey/convey"
)

// portalExpectedFiles is the file set generated for the portal namespace of the
// newsportal fixture, shared by every template set check.
var portalExpectedFiles = map[string]struct{}{
	filepath.Join("Category", "List.vue"):                           {},
	filepath.Join("Category", "Form.vue"):                           {},
	filepath.Join("Category", "en.json"):                            {},
	filepath.Join("Category", "components", "MultiListFilters.vue"): {},
	filepath.Join("News", "List.vue"):                               {},
	filepath.Join("News", "Form.vue"):                               {},
	filepath.Join("News", "en.json"):                                {},
	filepath.Join("News", "components", "MultiListFilters.vue"):     {},
	filepath.Join("Tag", "List.vue"):                                {},
	filepath.Join("Tag", "Form.vue"):                                {},
	filepath.Join("Tag", "en.json"):                                 {},
	filepath.Join("Tag", "components", "MultiListFilters.vue"):      {},
	"routes.ts": {},
}

// entityFilePrefix mirrors the layout the generator writes into the output dir
var entityFilePrefix = filepath.Join("src", "pages", "Entity")

// compareGenerated asserts that every file of the set is byte equal in the
// actual and the expected output dirs.
func compareGenerated(t *testing.T, actualPath, expectedPath string, files map[string]struct{}) {
	t.Helper()

	for f := range files {
		filenameWithFullPath := filepath.Join(actualPath, entityFilePrefix, f)
		t.Logf("Check %s file", filenameWithFullPath)
		content, err := os.ReadFile(filenameWithFullPath)
		So(err, ShouldBeNil)
		expectedContent, err := os.ReadFile(filepath.Join(expectedPath, entityFilePrefix, f))
		So(err, ShouldBeNil)
		So(string(content), ShouldResemble, string(expectedContent))
	}
}

func TestGenerator_Generate(t *testing.T) {
	// clean up leftover output from previous runs so partial routes injection starts fresh
	_ = os.RemoveAll(testdata.PathActualVTTemplateAll)
	_ = os.RemoveAll(testdata.PathActualVTTemplateEntity)

	Convey("TestGenerator_Generate", t, func() {
		generator := New()

		generator.options.Output = testdata.PathActualVTTemplateAll
		generator.options.MFDPath = testdata.PathExpectedMFD
		generator.options.Namespaces = []string{"portal"}
		// expected testdata is generated with default vue2 class-based templates
		// (project mfd has no VTTemplate setting)

		Convey("Check correct generate", func() {
			t.Log("Generate vt-template")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			compareGenerated(t, testdata.PathActualVTTemplateAll, testdata.PathExpectedVTTemplateAll, portalExpectedFiles)
		})

		Convey("Check correct generate with entities", func() {
			generator.options.Output = filepath.Join(testdata.PathActual, "vt-template", "entities")
			// two Entities
			generator.options.Entities = []string{"Category", "Tag"}

			t.Log("Generate vt-template with entities")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files with entities", func() {
			expectedFilenames := map[string]struct{}{
				filepath.Join("Category", "List.vue"):                           {},
				filepath.Join("Category", "Form.vue"):                           {},
				filepath.Join("Category", "en.json"):                            {},
				filepath.Join("Category", "components", "MultiListFilters.vue"): {},
				filepath.Join("Tag", "Form.vue"):                                {},
				filepath.Join("Tag", "List.vue"):                                {},
				filepath.Join("Tag", "en.json"):                                 {},
				filepath.Join("Tag", "components", "MultiListFilters.vue"):      {},
				"routes.ts": {},
			}

			Convey("Check content", func() {
				compareGenerated(t, testdata.PathActualVTTemplateEntity, testdata.PathExpectedVTTemplateEntity, expectedFilenames)
			})

			Convey("Check filenames", func() {
				actualFiles, err := fullFilesPaths(testdata.PathExpectedVTTemplateEntity)
				So(err, ShouldBeNil)

				for _, a := range actualFiles {
					shortPath := strings.ReplaceAll(a, filepath.Join(testdata.PathExpectedVTTemplateEntity, entityFilePrefix)+string(os.PathSeparator), "")
					t.Logf("Check %s filename", shortPath)
					_, ok := expectedFilenames[shortPath]
					So(ok, ShouldBeTrue)
				}
			})
		})
	})
}

// buildVTTemplateMFD writes a copy of the expected mfd with the project-level
// VTTemplate setting next to the original one, so translation and namespace
// files are still resolved, and returns its path.
func buildVTTemplateMFD(t *testing.T, vtTemplate string) string {
	t.Helper()

	project, err := mfd.LoadProject(testdata.PathExpectedMFD, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	project.VTTemplate = vtTemplate

	return saveTempMFD(t, filepath.Join(filepath.Dir(testdata.PathExpectedMFD), "newsportal."+vtTemplate+".mfd"), project)
}

// saveTempMFD writes project to mfdPath and removes it when the test ends. The
// path has to sit next to the fixture so namespace and translation files, which
// LoadProject resolves relative to the mfd, are still found.
func saveTempMFD(t *testing.T, mfdPath string, project *mfd.Project) string {
	t.Helper()

	if err := mfd.SaveMFD(mfdPath, project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(mfdPath) })

	return mfdPath
}

// checkVTTemplateSet generates the portal namespace with the given VTTemplate
// set and compares every generated file with the expected testdata.
func checkVTTemplateSet(t *testing.T, vtTemplate, actualPath, expectedPath string) {
	t.Helper()

	// clean up leftover output from previous runs so partial routes injection starts fresh
	_ = os.RemoveAll(actualPath)

	Convey("TestGenerator_Generate"+vtTemplate, t, func() {
		generator := New()

		generator.options.Output = actualPath
		// template set is selected via the project-level VTTemplate setting
		generator.options.MFDPath = buildVTTemplateMFD(t, vtTemplate)
		generator.options.Namespaces = []string{"portal"}

		Convey("Check correct generate with "+vtTemplate+" templates", func() {
			t.Logf("Generate vt-template with %s templates", vtTemplate)
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated "+vtTemplate+" files", func() {
			compareGenerated(t, actualPath, expectedPath, portalExpectedFiles)
		})
	})
}

func TestGenerator_GenerateComposition(t *testing.T) {
	checkVTTemplateSet(t, mfd.VTTemplateComposition, testdata.PathActualVTTemplateComposition, testdata.PathExpectedVTTemplateComposition)
}

func TestGenerator_GenerateVue3(t *testing.T) {
	checkVTTemplateSet(t, mfd.VTTemplateVue3, testdata.PathActualVTTemplateVue3, testdata.PathExpectedVTTemplateVue3)
}

func TestGenerator_UnknownVTTemplate(t *testing.T) {
	// expected vue2 testdata is generated from the mfd without VTTemplate, so an
	// unknown value must produce exactly the same files
	checkVTTemplateSet(t, "vue4", filepath.Join(testdata.PathActual, "vt-template", "unknown"), testdata.PathExpectedVTTemplateAll)
}

func TestGenerator_ModelAccessPrefix(t *testing.T) {
	dir := filepath.Dir(testdata.PathExpectedMFD)

	// buildMediaMFD writes a temp mfd next to the media/vfs namespace files so
	// LoadProject can resolve them; VTTemplate drives the access prefix.
	buildMediaMFD := func(vtTemplate string) string {
		p := mfd.NewProject("media.mfd", mfd.GoPG10)
		p.NamespaceNames = []string{"media", "vfs"}
		p.VTTemplate = vtTemplate

		return saveTempMFD(t, filepath.Join(dir, "media."+vtTemplate+".gen.mfd"), p)
	}

	generateForm := func(vtTemplate string) (string, error) {
		out := filepath.Join(testdata.PathActual, "vt-template", "media-"+vtTemplate)
		_ = os.RemoveAll(out)

		generator := New()
		generator.options.Output = out
		generator.options.MFDPath = buildMediaMFD(vtTemplate)
		generator.options.Namespaces = []string{"media"}
		if err := generator.Generate(); err != nil {
			return "", err
		}

		content, err := os.ReadFile(filepath.Join(out, "src", "pages", "Entity", "Page", "Form.vue"))
		return string(content), err
	}

	Convey("model access prefix depends on VTTemplate", t, func() {
		Convey("vue2 class-based templates use store.model.", func() {
			form, err := generateForm(mfd.VTTemplateVue2)
			So(err, ShouldBeNil)
			So(form, ShouldContainSubstring, `:value-for-transliterating="store.model.title"`)
			So(form, ShouldContainSubstring, `:file="store.model.`)
			So(form, ShouldNotContainSubstring, `:value-for-transliterating="model.`)
		})

		Convey("composition templates use model.", func() {
			form, err := generateForm(mfd.VTTemplateComposition)
			So(err, ShouldBeNil)
			So(form, ShouldContainSubstring, `:value-for-transliterating="model.title"`)
			So(form, ShouldContainSubstring, `:file="model.`)
			So(form, ShouldNotContainSubstring, `store.model.`)
		})

		Convey("vue3 templates use model. and the vuetify 3 display api", func() {
			form, err := generateForm(mfd.VTTemplateVue3)
			So(err, ShouldBeNil)
			So(form, ShouldContainSubstring, `:value-for-transliterating="model.title"`)
			So(form, ShouldContainSubstring, `:file="model.`)
			So(form, ShouldNotContainSubstring, `store.model.`)
			So(form, ShouldNotContainSubstring, `$vuetify.breakpoint.`)
		})
	})
}

// TestVue3ListValues covers the vue3 list cell expressions: Vue 3 has no
// filters, so the tableDate helper is called directly (and imported/returned
// only when it is really used) and a foreign key field is read from the nested
// object. The newsportal testdata has no date column in a list, so the template
// is rendered here directly.
func TestVue3ListValues(t *testing.T) {
	render := func(columns []AttributeData) string {
		parsed, err := parseEntityTemplate(listVue3Template)
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		data := EntityData{
			Name:        "News",
			JSName:      "news",
			PKs:         []PKPair{{JSName: "id"}},
			ListColumns: columns,
		}
		if err := parsed.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatal(err)
		}

		return buf.String()
	}

	Convey("vue3 list renders cell values without vue 2 filters", t, func() {
		Convey("date column calls tableDate and imports it", func() {
			out := render([]AttributeData{
				{JSName: "createdAt", HasPipe: true, Pipe: "tableDate", IsTableDate: true, Value: "tableDate(item.createdAt)", IsSortable: true},
			})

			So(out, ShouldContainSubstring, "{{ tableDate(item.createdAt) }}")
			So(out, ShouldContainSubstring, "import { tableDate } from '@/helpers/date';")
			So(out, ShouldContainSubstring, "\n      tableDate,\n")
			So(out, ShouldNotContainSubstring, "|")
		})

		Convey("fk column reads the nested field, no tableDate import", func() {
			out := render([]AttributeData{
				{JSName: "category", HasPipe: true, Pipe: `getField("title")`, Value: "item.category?.title"},
			})

			So(out, ShouldContainSubstring, "{{ item.category?.title }}")
			So(out, ShouldNotContainSubstring, "getField")
			So(out, ShouldNotContainSubstring, "@/helpers/date")
			So(out, ShouldNotContainSubstring, "tableDate")
		})
	})
}

func fullFilesPaths(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var filePaths []string
	for _, file := range files {
		if file.IsDir() {
			paths, err := fullFilesPaths(filepath.Join(path, file.Name()))
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, paths...)
		} else {
			filePaths = append(filePaths, filepath.Join(path, file.Name()))
		}
	}

	return filePaths, nil
}

func TestManualGenerate(t *testing.T) {
	t.Skip()
	generator := New()

	generator.options.Output = ""
	generator.options.MFDPath = ""
	generator.options.Namespaces = []string{"article", "catalogue"}
	generator.options.Entities = []string{"tag", "category"}

	err := generator.Generate()
	if err != nil {
		t.Error(err)
	}
}

func Test_extractEntityBlock(t *testing.T) {
	type args struct {
		content    string
		entityName string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "middle body contains",
			args: args{
				entityName: "Category",
				content: `export default [
  
  /* News */
  {
    name: "newsList",
    path: "/news",
  },
  /* Category */
  {
    name: "categoryList",
    path: "/category",
  },
  /* Tag */
  {
    name: "tagList",
  },
];`,
			},
			want: []string{
				`{`,
				`  name: "categoryList",`,
				`  path: "/category",`,
				`},`,
			},
		},
		{
			name: "last body",
			args: args{
				entityName: "Tag",
				content: `export default [
  /* Category */
  {
    name: "categoryList",
  },
  /* Tag */
  {
    name: "tagList",
    path: "/tags",
  }
];`,
			},
			want: []string{
				`{`,
				`  name: "tagList",`,
				`  path: "/tags",`,
				`}`,
			},
		},
		{
			name: "now found",
			args: args{
				entityName: "News",
				content: `export default [
  /* Category */
  {
    name: "categoryList",
  },
];`,
			},
			want: nil,
		},
		{
			name: "empty content",
			args: args{
				entityName: "Category",
				content:    ``,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEntityBlock(tt.args.content, tt.args.entityName)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractEntityBlock() =\n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}

func Test_injectEntityBlock(t *testing.T) {
	sampleNewBlock := []string{
		`{`,
		`  name: "tagList",`,
		`  path: "/tags",`,
		`},`,
	}

	type args struct {
		existingContent string
		entityName      string
		newBlock        []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "update middle body",
			args: args{
				entityName: "Tag",
				newBlock:   sampleNewBlock,
				existingContent: `export default [
  /* Category */
  {
    name: "categoryList",
  },
  /* Tag */
  {
    name: "oldTagList",
  },
  /* News */
  {
    name: "newsList",
  }
];`,
			},
			want: `export default [
  /* Category */
  {
    name: "categoryList",
  },
    /* Tag */
  {
    name: "tagList",
    path: "/tags",
  },
  /* News */
  {
    name: "newsList",
  }
];`,
		},
		{
			name: "insert new data variant 1",
			args: args{
				entityName: "Tag",
				newBlock:   sampleNewBlock,
				existingContent: `export default [
  /* Category */
  {
    name: "categoryList",
  },
];`,
			},
			want: `export default [
  /* Category */
  {
    name: "categoryList",
  },
    /* Tag */
  {
    name: "tagList",
    path: "/tags",
  },
];`,
		},
		{
			name: "insert new data variant 2",
			args: args{
				entityName: "Tag",
				newBlock:   sampleNewBlock,
				existingContent: `export default [
  /* Category */
  {
    name: "categoryList"
  }
];`,
			},
			want: `export default [
  /* Category */
  {
    name: "categoryList"
  },
    /* Tag */
  {
    name: "tagList",
    path: "/tags",
  },
];`,
		},
		{
			name: "insert new data variant 2",
			args: args{
				entityName: "Tag",
				newBlock:   sampleNewBlock,
				existingContent: `export default [
];`,
			},
			want: `export default [
    /* Tag */
  {
    name: "tagList",
    path: "/tags",
  },
];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectEntityBlock(tt.args.existingContent, tt.args.entityName, tt.args.newBlock)
			if got != tt.want {
				t.Errorf("injectEntityBlock() failed.\n\n GOT \n%s\n\n WANT \n%s\n", got, tt.want)
			}
		})
	}
}
