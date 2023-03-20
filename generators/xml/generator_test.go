package xml

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmkteam/mfd-generator/generators/testdata"
)

// todo: fail if ../testdata/necessary-content/actual/*.mfd,*.xml exists,
// todo: panic if ../testdata/necessary-content/actual/*.mfd exists and xml not exist
func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = `postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable`
			generator.options.Output = testdata.PathActualMfd
			generator.options.Packages = parseNamespacesFlag("portal:news,categories,tags")

			t.Log("Generate xml")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.xml":     {},
				"newsportal.mfd": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(testdata.PathActual + f)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(testdata.PathExpected + f)
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
