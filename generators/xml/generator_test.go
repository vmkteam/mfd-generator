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

// TestDecideNamespace checks the pure namespace-resolution logic that drives
// --quiet (-q) and --namespaces (-n). It replaces the previous DB-backed test
// that swapped an interactive-prompt seam on the Generator: every invariant the
// seam used to verify (which tables are mapped, skipped or prompted, and that -n
// overrides -q/TableMapping) is now asserted directly, with no DB and no mocks.
func TestDecideNamespace(t *testing.T) {
	// tableMapping mirrors a project where news/categories/tags are pre-mapped to "portal".
	tableMapping := map[string]string{
		"news":       "portal",
		"categories": "portal",
		"tags":       "portal",
	}

	tests := []struct {
		name        string
		opts        Options
		table       string
		existingNS  string
		hasExisting bool
		wantNS      string
		wantAction  nsAction
	}{
		{
			name:       "-q new: mapped table is assigned silently",
			opts:       Options{Quiet: quietNew},
			table:      "news",
			wantNS:     "portal",
			wantAction: nsAssign,
		},
		{
			name:        "-q new: unmapped table already in project keeps its namespace",
			opts:        Options{Quiet: quietNew},
			table:       "comments",
			existingNS:  "blog",
			hasExisting: true,
			wantNS:      "blog",
			wantAction:  nsAssign,
		},
		{
			name:       "-q new: brand new table must be prompted",
			opts:       Options{Quiet: quietNew},
			table:      "vfsFiles",
			wantAction: nsPrompt,
		},
		{
			name:       "-q all: mapped table is assigned",
			opts:       Options{Quiet: quietAll},
			table:      "news",
			wantNS:     "portal",
			wantAction: nsAssign,
		},
		{
			name:        "-q all: unmapped table already in project keeps its namespace",
			opts:        Options{Quiet: quietAll},
			table:       "comments",
			existingNS:  "blog",
			hasExisting: true,
			wantNS:      "blog",
			wantAction:  nsAssign,
		},
		{
			name:       "-q all: unmapped new table is skipped",
			opts:       Options{Quiet: quietAll},
			table:      "vfsFolders",
			wantAction: nsSkip,
		},
		{
			name:       "-n: listed table is assigned",
			opts:       Options{Packages: map[string]string{"encryptionKeys": "custom"}},
			table:      "encryptionKeys",
			wantNS:     "custom",
			wantAction: nsAssign,
		},
		{
			name:       "-n: table missing from preset is skipped",
			opts:       Options{Packages: map[string]string{"encryptionKeys": "custom"}},
			table:      "vfsFolders",
			wantAction: nsSkip,
		},
		{
			name:       "-n overrides TableMapping",
			opts:       Options{Quiet: quietNew, Packages: map[string]string{"news": "custom"}},
			table:      "news",
			wantNS:     "custom",
			wantAction: nsAssign,
		},
		{
			name:       "default (no -q/-n): new table is prompted",
			opts:       Options{},
			table:      "news",
			wantAction: nsPrompt,
		},
		{
			name:        "default (no -q/-n): even an existing table is prompted",
			opts:        Options{},
			table:       "comments",
			existingNS:  "blog",
			hasExisting: true,
			wantAction:  nsPrompt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNS, gotAction := decideNamespace(tt.opts, tableMapping, tt.table, tt.existingNS, tt.hasExisting)
			if gotAction != tt.wantAction {
				t.Errorf("action = %d, want %d", gotAction, tt.wantAction)
			}
			if gotNS != tt.wantNS {
				t.Errorf("namespace = %q, want %q", gotNS, tt.wantNS)
			}
		})
	}
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
