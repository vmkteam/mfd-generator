package xml

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"
	"github.com/vmkteam/mfd-generator/mfd"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: panic if ../testdata/actual/*.mfd exists and xml not exist
func TestGenerator_Generate(t *testing.T) {
	// Store the PATH environment variable in a variable
	actualDir := t.TempDir()
	mfdPathInActual := filepath.Join(actualDir, filepath.Base(testdata.PathExpectedMFD))

	dbdsn, exists := os.LookupEnv("DB_DSN")
	if !exists {
		dbdsn = "postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable"
	}

	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.URL = dbdsn
			generator.options.Output = mfdPathInActual
			generator.options.Packages = parseNamespacesFlag("portal:news,categories,tags")

			t.Log("Generate xml")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.xml":     {},
				"newsportal.mfd": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(actualDir, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpected, f))
				if err != nil {
					t.Fatal(err)
				}
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}

func helperLoadBytes(t *testing.T, path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestGenerator_TableModel(t *testing.T) {
	actualDir := t.TempDir()
	mfdPathInActual := filepath.Join(actualDir, filepath.Base(testdata.PathExpectedMFD))

	Convey("TestGenerator_TableModel", t, func() {
		Convey("Check equal generate tableModels", func() {
			var project, projectActual mfd.Project
			err := xml.Unmarshal(helperLoadBytes(t, testdata.PathExpectedMFD), &project)
			So(err, ShouldBeNil)
			project.TableMapping = mfd.TableMapping{
				Entries: []mfd.Entry{
					{
						XMLName: xml.Name{Local: "portal"},
						Value:   "news,categories,tags",
					},
				},
			}
			err = mfd.SaveMFD(mfdPathInActual, &project)
			So(err, ShouldBeNil)

			err = xml.Unmarshal(helperLoadBytes(t, mfdPathInActual), &projectActual)
			So(err, ShouldBeNil)
			So(reflect.DeepEqual(projectActual.TableMapping, project.TableMapping), ShouldBeTrue)
		})
		Convey("Check equal Packages data", func() {
			generator := New()

			generator.options.Def()
			generator.options.Packages = parseNamespacesFlag("common:users;vfs:vfsFiles,vfsFolders;news:news,categories,tags")

			project := mfd.Project{TableMapping: mfd.TableMapping{
				Entries: []mfd.Entry{
					{
						XMLName: xml.Name{Local: "common"},
						Value:   "users",
					},
					{
						XMLName: xml.Name{Local: "vfs"},
						Value:   "vfsFiles,vfsFolders",
					},
					{
						XMLName: xml.Name{Local: "news"},
						Value:   "news,categories,tags",
					},
				},
			}}

			project.TableMapping.Packages()

			So(reflect.DeepEqual(generator.options.Packages, project.TableMapping.Packages()), ShouldBeTrue)
		})
	})
}
