package xml

import (
	"os"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = `postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable`
			generator.options.Output = "result/newsportal.mfd"
			generator.options.Packages = parseNamespacesFlag("common:users;vfs:vfsFiles,vfsFolders;portal:news,categories,tags")

			err := generator.Generate()
			So(err, ShouldBeNil)
		})
		Convey("Check generated files", func() {
			files, err := os.ReadDir("testdata/result/")
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range files {
				content, err := os.ReadFile("testdata/result/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				necessaryContent, err := os.ReadFile("testdata/necessary/" + f.Name())
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, necessaryContent)
			}
		})
	})
}
