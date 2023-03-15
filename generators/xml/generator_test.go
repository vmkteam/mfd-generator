package xml

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: fail if testdata/necessaryContent/actual/*.mfd,*.xml exists,
// todo: panic if testdata/necessaryContent/actual/*.mfd exists and xml not exist
func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = `postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable`
			generator.options.Output = "testdata/necessaryContent/actual/newsportal.mfd"
			generator.options.Packages = parseNamespacesFlag("common:users;vfs:vfsFiles,vfsFolders;portal:news,categories,tags")

			err := generator.Generate()
			So(err, ShouldBeNil)
		})
		Convey("Check generated files", func() {
			files, err := os.ReadDir("testdata/necessaryContent/actual/")
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range files {
				content, err := os.ReadFile("testdata/necessaryContent/actual/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile("testdata/necessaryContent/expected/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
