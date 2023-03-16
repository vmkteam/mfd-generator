package repo

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
			generator.options.Output = testdata.PathActualDB
			generator.options.MFDPath = testdata.PathExpectedMfd
			generator.options.Package = testdata.PackageDB
			generator.options.Namespaces = []string{"portal"}

			t.Log("Generate repo")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.go": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(testdata.PathActualDB + f)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(testdata.PathExpectedDB + f)
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
