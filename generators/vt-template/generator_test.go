package vttmpl

import (
	"os"
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		generator := New()

		generator.options.Output = testdata.PathActualVtTemplateAll
		generator.options.MFDPath = testdata.PathExpectedMfd
		generator.options.Namespaces = []string{"portal"}

		Convey("Check correct generate", func() {
			t.Log("Generate vt-template")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		filePrefix := "src/pages/Entity/"

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"Category/List.vue":                        {},
				"Category/components/MultiListFilters.vue": {},
				"News/Form.vue":                            {},
				"News/List.vue":                            {},
				"Tag/Form.vue":                             {},
				"Tag/List.vue":                             {},
				"Category/Form.vue":                        {},
				"News/components/MultiListFilters.vue":     {},
				"Tag/components/MultiListFilters.vue":      {},
				"routes.ts":                                {},
			}

			for f := range expectedFilenames {
				filenameWithFullPath := testdata.PathActualVtTemplateAll + filePrefix + f
				t.Logf("Check %s file", filenameWithFullPath)
				content, err := os.ReadFile(filenameWithFullPath)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(testdata.PathExpectedVtTemplateAll + filePrefix + f)
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})

		Convey("Check correct generate with entities", func() {
			generator.options.Output = testdata.PathActual + "vt-template/entities/"
			generator.options.Entities = []string{"Category", "Tag"}

			t.Log("Generate vt-template with entities")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files with entities", func() {
			expectedFilenames := map[string]struct{}{
				"Category/List.vue":                        {},
				"Category/components/MultiListFilters.vue": {},
				"Tag/Form.vue":                             {},
				"Tag/List.vue":                             {},
				"Category/Form.vue":                        {},
				"Tag/components/MultiListFilters.vue":      {},
				"routes.ts":                                {},
			}

			Convey("Check content", func() {
				for f := range expectedFilenames {
					filenameWithFullPath := testdata.PathActualVtTemplateEntity + filePrefix + f
					t.Logf("Check %s file", filenameWithFullPath)
					content, err := os.ReadFile(filenameWithFullPath)
					if err != nil {
						t.Fatal(err)
					}
					expectedContent, err := os.ReadFile(testdata.PathExpectedVtTemplateEntity + filePrefix + f)
					if err != nil {
						t.Fatal(err)
					}
					So(content, ShouldResemble, expectedContent)
				}
			})

			Convey("Check filenames", func() {
				actualFiles, err := fullFilesPaths(testdata.PathExpectedVtTemplateEntity)
				if err != nil {
					t.Fatal(err)
				}

				for _, a := range actualFiles {
					shortPath := strings.ReplaceAll(a, testdata.PathExpectedVtTemplateEntity+filePrefix, "")
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
			paths, err := fullFilesPaths(path + file.Name() + "/")
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, paths...)
		} else {
			filePaths = append(filePaths, path+file.Name())
		}
	}

	return filePaths, nil
}
