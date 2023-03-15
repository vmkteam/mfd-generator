package xml

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: fail if ../testdata/necessary-content/actual/*.mfd,*.xml exists,
// todo: panic if ../testdata/necessary-content/actual/*.mfd exists and xml not exist
func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = `postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable`
			generator.options.Output = "../testdata/necessary-content/actual/newsportal.mfd"
			generator.options.Packages = parseNamespacesFlag("portal:news,categories,tags")

			err := generator.Generate()
			So(err, ShouldBeNil)
		})
		Convey("Check generated files", func() {
			files, err := os.ReadDir("../testdata/necessary-content/actual/")
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range files {
				content, err := os.ReadFile("../testdata/necessary-content/actual/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile("../testdata/necessary-content/expected/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
