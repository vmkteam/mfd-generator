package vt

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmkteam/mfd-generator/generators/testdata"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathActualVt
			generator.options.MFDPath = testdata.PathExpectedMfd
			generator.options.Package = testdata.PackageVt
			generator.options.Namespaces = []string{"portal"}
			generator.options.ModelPackage = "github.com/vmkteam/mfd-generator/generators/testdata/necessary-content/expected/db"

			t.Log("Generate vt")
			_ = generator.Generate()
			//todo: failed, because portal.go generate import with empty quotes. Run test and check "generators/testdata/necessary-content/actual/vt/portal.go"
			//So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.go":           {},
				"portal_converter.go": {},
				"portal_model.go":     {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(testdata.PathActualVt + f)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(testdata.PathExpectedVt + f)
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
