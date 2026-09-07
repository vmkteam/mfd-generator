package dbtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOptions_DefDefaultsToLegacyRepoMode(t *testing.T) {
	var options Options

	options.Def()

	if options.RepoMode != repo.ModeLegacy {
		t.Fatalf("RepoMode = %q, want %q", options.RepoMode, repo.ModeLegacy)
	}
}

func TestGenerator_Generate(t *testing.T) {
	t.Run("Without special conditions", func(t *testing.T) {
		actualDir := t.TempDir()
		Convey("Without special conditions", t, func() {
			Convey("Check correct generate", func() {
				generator := New()
				generator.options.Def()
				generator.options.Output = actualDir
				generator.options.MFDPath = testdata.PathExpectedMFD
				generator.options.Package = testdata.PackageDBTest
				generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

				t.Log("Generate model")
				So(generator.Generate(), ShouldBeNil)
			})

			Convey("Check generated files", func() {
				expectedFilenames := map[string]struct{}{
					"test.go":   {},
					"geo.go":    {},
					"portal.go": {},
					"vfs.go":    {},
				}

				for f := range expectedFilenames {
					t.Logf("Check %s file", f)
					content, err := os.ReadFile(filepath.Join(actualDir, f))
					if err != nil {
						t.Fatal(err)
					}
					expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedDBTest, f))
					if err != nil {
						t.Fatal(err)
					}

					So(string(content), ShouldResemble, string(expectedContent))
				}
			})
		})
	})
	t.Run("Portal namespace only", func(t *testing.T) {
		actualDir := t.TempDir()
		Convey("Portal namespace only", t, func() {
			Convey("Check correct generate", func() {
				generator := New()
				generator.options.Def()
				generator.options.Output = actualDir
				generator.options.MFDPath = testdata.PathExpectedMFD
				generator.options.Package = testdata.PackageDBTest
				generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/actual/db"
				generator.options.Namespaces = []string{"portal"}

				t.Log("Generate model")
				So(generator.Generate(), ShouldBeNil)
			})

			Convey("Check generated files", func() {
				expectedFilenames := map[string]struct{}{
					"test.go":   {},
					"portal.go": {},
				}

				for f := range expectedFilenames {
					t.Logf("Check %s file", f)
					content, err := os.ReadFile(filepath.Join(actualDir, f))
					if err != nil {
						t.Fatal(err)
					}
					expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedDBTest, f))
					if err != nil {
						t.Fatal(err)
					}

					So(string(content), ShouldResemble, string(expectedContent))
				}

				mustNotBeCreatedFilenames := map[string]struct{}{
					"geo.go": {},
				}
				for f := range mustNotBeCreatedFilenames {
					_, err := os.Stat(filepath.Join(actualDir, f))
					So(os.IsNotExist(err), ShouldBeTrue)
				}
			})
		})
	})
	t.Run("News entity only", func(t *testing.T) {
		actualDir := t.TempDir()
		Convey("News entity only", t, func() {
			Convey("Check correct generate", func() {
				generator := New()
				generator.options.Def()
				generator.options.Output = actualDir
				generator.options.MFDPath = testdata.PathExpectedMFD
				generator.options.Package = testdata.PackageDBTest
				generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/actual/db"
				generator.options.Entities = []string{"news"}

				t.Log("Generate model")
				So(generator.Generate(), ShouldBeNil)
			})

			Convey("Check generated files", func() {
				filename := "portal.go"
				t.Logf("Check %s file", filename)
				res, err := funcNamesInFile(filepath.Join(actualDir, filename))
				So(err, ShouldBeNil)
				So(res, ShouldResemble, []string{"News", "WithNewsRelations", "WithFakeNews"})
			})
		})
	})
	t.Run("News entity with the namespace and Region entity without the namespace", func(t *testing.T) {
		actualDir := t.TempDir()
		Convey("News entity with the namespace and Region entity without the namespace", t, func() {
			Convey("Check correct generate", func() {
				generator := New()
				generator.options.Def()
				generator.options.Output = actualDir
				generator.options.MFDPath = testdata.PathExpectedMFD
				generator.options.Package = testdata.PackageDBTest
				generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/actual/db"
				generator.options.Entities = []string{"news", "region"}
				generator.options.Namespaces = []string{"portal"}

				t.Log("Generate model")
				So(generator.Generate(), ShouldBeNil)
			})

			Convey("Check generated files", func() {
				filename := "portal.go"
				t.Logf("Check %s file", filename)
				res, err := funcNamesInFile(filepath.Join(actualDir, filename))
				So(err, ShouldBeNil)
				So(res, ShouldResemble, []string{"News", "WithNewsRelations", "WithFakeNews"})
				filename = "geo.go"
				t.Logf("Check %s file", filename)
				res, err = funcNamesInFile(filepath.Join(actualDir, filename))
				So(err, ShouldBeNil)
				So(res, ShouldResemble, []string{"Region", "WithRegionRelations", "WithFakeRegion"})
			})
		})
	})
	t.Run("Force all", func(t *testing.T) {
		actualDir := t.TempDir()
		Convey("News entity only", t, func() {
			filename := "portal.go"
			Convey("First generating", func() {
				generator := New()
				generator.options.Def()
				generator.options.Output = actualDir
				generator.options.MFDPath = testdata.PathExpectedMFD
				generator.options.Package = testdata.PackageDBTest
				generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

				t.Log("Generate model")
				So(generator.Generate(), ShouldBeNil)

				// Change func content
				content, err := os.ReadFile(filepath.Join(actualDir, filename))
				So(err, ShouldBeNil)

				const oldContent = `// Create the main entity
	news, err := repo.AddNews(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}`
				newContent := strings.ReplaceAll(string(content), oldContent, "")
				err = os.WriteFile(filepath.Join(actualDir, filename), []byte(newContent), 0644)
				So(err, ShouldBeNil)

				generator.options.Entities = []string{"news"}
				generator.options.Force = true
				So(generator.Generate(), ShouldBeNil)
			})

			Convey("Check generated files", func() {
				t.Logf("Check %s file", filename)
				content, err := os.ReadFile(filepath.Join(actualDir, filename))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedDBTest, filename))
				if err != nil {
					t.Fatal(err)
				}

				So(string(content), ShouldResemble, string(expectedContent))
			})
		})
	})
}

func funcNamesInFile(output string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, output, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file, path=%s, err=%w", output, err)
	}

	var res []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			res = append(res, fn.Name.Name)
		}
		return true
	})

	return res, nil
}
