package xml

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	. "github.com/smartystreets/goconvey/convey"
)

// todo: error if .mfd file has already existed but .xml files was deleted
func TestGenerator_Generate(t *testing.T) {
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
			generator.options.CustomTypes = model.CustomTypeMapping{"uuid": {
				PGType:   "uuid",
				GoType:   "uuid.UUID",
				GoImport: "github.com/google/uuid",
			}}
			generator.options.Packages = parseNamespacesFlag("portal:news,categories,tags;geo:countries,regions,cities;vfs:vfsFiles,vfsFolders;card:encryptionKeys;common:siteUsers,loginCodes")

			t.Log("Generate xml")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.xml":     {},
				"geo.xml":        {},
				"card.xml":       {},
				"common.xml":     {},
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
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}

// TestGenerator_QuietWithTableMapping checks that --quiet (-q) flag is respected
// when -n is not set but the project already has a TableMapping section.
// Help promises that -q is ignored only when -n is set; with -n absent,
// -q new must still prompt for tables that are missing from the mapping.
func TestGenerator_QuietWithTableMapping(t *testing.T) {
	dbdsn, exists := os.LookupEnv("DB_DSN")
	if !exists {
		dbdsn = "postgres://postgres:postgres@localhost:5432/newsportal?sslmode=disable"
	}

	customTypes := model.CustomTypeMapping{"uuid": {
		PGType:   "uuid",
		GoType:   "uuid.UUID",
		GoImport: "github.com/google/uuid",
	}}

	seedMFD := func(t *testing.T, entries ...mfd.Entry) string {
		t.Helper()
		mfdPath := filepath.Join(t.TempDir(), testdata.FilenameMFD)
		project := mfd.NewProject(testdata.FilenameMFD, mfd.GoPG10)
		project.TableMapping = mfd.TableMapping{Entries: entries}
		So(mfd.SaveMFD(mfdPath, project), ShouldBeNil)
		return mfdPath
	}

	// fakePrompt mirrors PromptNS behaviour for special tables but routes the rest
	// to the supplied namespace, recording every call.
	fakePrompt := func(target string, prompted map[string]struct{}) func(string, []string) (string, error) {
		return func(table string, _ []string) (string, error) {
			prompted[table] = struct{}{}
			if table == "statuses" {
				return "skip", nil
			}
			return target, nil
		}
	}

	Convey("TestGenerator_QuietWithTableMapping", t, func() {
		Convey("-q new prompts only for tables outside TableMapping", func() {
			mfdPath := seedMFD(t, mfd.Entry{
				XMLName: xml.Name{Local: "portal"},
				Value:   "news,categories,tags",
			})

			prompted := map[string]struct{}{}
			generator := New()
			generator.options.URL = dbdsn
			generator.options.Output = mfdPath
			generator.options.GoPgVer = mfd.GoPG10
			generator.options.Quiet = quietNew
			generator.options.CustomTypes = customTypes
			generator.promptNS = fakePrompt("other", prompted)

			t.Log("Generate xml with -q new and a pre-seeded TableMapping")
			So(generator.Generate(), ShouldBeNil)

			// mapped tables must be assigned silently — no prompt
			So(prompted, ShouldNotContainKey, "news")
			So(prompted, ShouldNotContainKey, "categories")
			So(prompted, ShouldNotContainKey, "tags")
			// new tables must trigger the prompt
			So(prompted, ShouldContainKey, "vfsFiles")
			So(prompted, ShouldContainKey, "countries")

			generated, err := mfd.LoadProject(mfdPath, false, mfd.GoPG10)
			So(err, ShouldBeNil)

			portal := generated.Namespace("portal")
			So(portal, ShouldNotBeNil)
			So(portal.EntityByTable("news"), ShouldNotBeNil)
			So(portal.EntityByTable("categories"), ShouldNotBeNil)
			So(portal.EntityByTable("tags"), ShouldNotBeNil)

			other := generated.Namespace("other")
			So(other, ShouldNotBeNil)
			So(other.EntityByTable("vfsFiles"), ShouldNotBeNil)
			So(other.EntityByTable("countries"), ShouldNotBeNil)
		})

		Convey("-q all with TableMapping keeps mapping and skips unmapped tables", func() {
			// encryptionKeys has no FK on other regular entities (only on the statuses
			// enum-table), so it can be mapped on its own without breaking IsConsistent;
			// vfsFolders is read but stays unmapped to exercise the skip path.
			mfdPath := seedMFD(t, mfd.Entry{
				XMLName: xml.Name{Local: "card"},
				Value:   "encryptionKeys",
			})

			promptCalls := 0
			generator := New()
			generator.options.URL = dbdsn
			generator.options.Output = mfdPath
			generator.options.Tables = []string{"public.encryptionKeys", "public.vfsFolders"}
			generator.options.GoPgVer = mfd.GoPG10
			generator.options.Quiet = quietAll
			generator.options.CustomTypes = customTypes
			generator.promptNS = func(string, []string) (string, error) {
				promptCalls++
				return "", nil
			}

			t.Log("Generate xml with -q all and a pre-seeded TableMapping")
			So(generator.Generate(), ShouldBeNil)
			So(promptCalls, ShouldEqual, 0)

			generated, err := mfd.LoadProject(mfdPath, false, mfd.GoPG10)
			So(err, ShouldBeNil)

			card := generated.Namespace("card")
			So(card, ShouldNotBeNil)
			So(card.EntityByTable("encryptionKeys"), ShouldNotBeNil)

			// unmapped table must not leak into the project under -q all
			So(generated.EntityByTable("vfsFolders"), ShouldBeNil)
		})

		Convey("-n preset still ignores -q (help contract)", func() {
			// seed TableMapping that points encryptionKeys at "portal" — if -q won
			// over -n, encryptionKeys would land in "portal".
			mfdPath := seedMFD(t, mfd.Entry{
				XMLName: xml.Name{Local: "portal"},
				Value:   "encryptionKeys",
			})

			promptCalls := 0
			generator := New()
			generator.options.URL = dbdsn
			generator.options.Output = mfdPath
			generator.options.GoPgVer = mfd.GoPG10
			generator.options.Quiet = quietNew
			generator.options.CustomTypes = customTypes
			// -n must take precedence and place encryptionKeys into "custom"
			generator.options.Packages = parseNamespacesFlag("custom:encryptionKeys")
			generator.promptNS = func(string, []string) (string, error) {
				promptCalls++
				return "", nil
			}

			t.Log("Generate xml with -n set and -q new together")
			So(generator.Generate(), ShouldBeNil)
			So(promptCalls, ShouldEqual, 0)

			generated, err := mfd.LoadProject(mfdPath, false, mfd.GoPG10)
			So(err, ShouldBeNil)

			custom := generated.Namespace("custom")
			So(custom, ShouldNotBeNil)
			So(custom.EntityByTable("encryptionKeys"), ShouldNotBeNil)
			// TableMapping must be overridden by -n
			So(generated.Namespace("portal"), ShouldBeNil)
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
