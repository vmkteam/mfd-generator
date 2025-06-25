package vt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathActualVT
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageVT
			generator.options.Namespaces = []string{"portal"}
			generator.options.ModelPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
			generator.options.EmbedLogPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/embedlog"

			t.Log("Generate vt")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.go":           {},
				"portal_converter.go": {},
				"portal_model.go":     {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(testdata.PathActualVT, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVT, f))
				if err != nil {
					t.Fatal(err)
				}
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}
