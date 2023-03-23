package vttmpl

import (
	"os"
	"path/filepath"
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
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		filePrefix := filepath.Join("src", "pages", "Entity")

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				filepath.Join("Category", "List.vue"):                           {},
				filepath.Join("Category", "components", "MultiListFilters.vue"): {},
				filepath.Join("News", "Form.vue"):                               {},
				filepath.Join("News", "List.vue"):                               {},
				filepath.Join("Tag", "Form.vue"):                                {},
				filepath.Join("Tag", "List.vue"):                                {},
				filepath.Join("Category", "Form.vue"):                           {},
				filepath.Join("News", "components", "MultiListFilters.vue"):     {},
				filepath.Join("Tag", "components", "MultiListFilters.vue"):      {},
				filepath.Join("routes.ts"):                                      {},
			}

			for f := range expectedFilenames {
				filenameWithFullPath := filepath.Join(testdata.PathActualVTTemplateAll, filePrefix, f)
				t.Logf("Check %s file", filenameWithFullPath)
				content, err := os.ReadFile(filenameWithFullPath)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVTTemplateAll, filePrefix, f))
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
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
				filepath.Join("Category", "components", "MultiListFilters.vue"): {},
				filepath.Join("Tag", "Form.vue"):                                {},
				filepath.Join("Tag", "List.vue"):                                {},
				filepath.Join("Category", "Form.vue"):                           {},
				filepath.Join("Tag", "components", "MultiListFilters.vue"):      {},
				filepath.Join("routes.ts"):                                      {},
			}

			Convey("Check content", func() {
				for f := range expectedFilenames {
					filenameWithFullPath := filepath.Join(testdata.PathActualVTTemplateEntity, filePrefix, f)
					t.Logf("Check %s file", filenameWithFullPath)
					content, err := os.ReadFile(filenameWithFullPath)
					if err != nil {
						t.Fatal(err)
					}
					expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVTTemplateEntity, filePrefix, f))
					if err != nil {
						t.Fatal(err)
					}
					So(content, ShouldResemble, expectedContent)
				}
			})

			Convey("Check filenames", func() {
				actualFiles, err := fullFilesPaths(testdata.PathExpectedVTTemplateEntity)
				if err != nil {
					t.Fatal(err)
				}

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
