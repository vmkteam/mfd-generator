package xml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: fail if ../testdata/actual/*.mfd,*.xml exists,
// todo: panic if ../testdata/actual/*.mfd exists and xml not exist
func TestGenerator_Generate(t *testing.T) {
	// Store the PATH environment variable in a variable
	dbdsn, exists := os.LookupEnv("DB_DSN")
	if !exists {
		dbdsn = "postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable"
	}

	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = dbdsn
			generator.options.Output = testdata.PathActualMFD
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
				content, err := os.ReadFile(filepath.Join(testdata.PathActual, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpected, f))
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
