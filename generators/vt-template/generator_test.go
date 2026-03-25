package vttmpl

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		generator := New()

		generator.options.Output = testdata.PathActualVTTemplateAll
		generator.options.MFDPath = testdata.PathExpectedMFD
		generator.options.Namespaces = []string{"portal"}

		Convey("Check correct generate", func() {
			t.Log("Generate vt-template")
			So(generator.Generate(), ShouldBeNil)
		})

		filePrefix := filepath.Join("src", "pages", "Entity")

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
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

			for f := range expectedFilenames {
				filenameWithFullPath := filepath.Join(testdata.PathActualVTTemplateAll, filePrefix, f)
				t.Logf("Check %s file", filenameWithFullPath)
				content, err := os.ReadFile(filenameWithFullPath)
				So(err, ShouldBeNil)
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVTTemplateAll, filePrefix, f))
				So(err, ShouldBeNil)
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})

		Convey("Check correct generate with entities", func() {
			generator.options.Output = filepath.Join(testdata.PathActual, "vt-template", "entities")
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
				for f := range expectedFilenames {
					filenameWithFullPath := filepath.Join(testdata.PathActualVTTemplateEntity, filePrefix, f)
					t.Logf("Check %s file", filenameWithFullPath)
					content, err := os.ReadFile(filenameWithFullPath)
					So(err, ShouldBeNil)
					expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVTTemplateEntity, filePrefix, f))
					So(err, ShouldBeNil)
					So(string(content), ShouldResemble, string(expectedContent))
				}
			})

			Convey("Check filenames", func() {
				actualFiles, err := fullFilesPaths(testdata.PathExpectedVTTemplateEntity)
				So(err, ShouldBeNil)

				for _, a := range actualFiles {
					shortPath := strings.ReplaceAll(a, filepath.Join(testdata.PathExpectedVTTemplateEntity, filePrefix)+string(os.PathSeparator), "")
					t.Logf("Check %s filename", shortPath)
					_, ok := expectedFilenames[shortPath]
					So(ok, ShouldBeTrue)
				}
			})
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
