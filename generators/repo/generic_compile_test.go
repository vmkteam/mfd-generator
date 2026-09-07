package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"
)

func TestGenericGeneratedCodeCompiles(t *testing.T) {
	output, err := filepath.Abs(filepath.Join(testdata.PathActual, "generic-compile-"+strconv.Itoa(os.Getpid())))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(output) })

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	mfdPath, err := filepath.Abs(testdata.PathExpectedMFD)
	if err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		cmd := exec.Command("go", args...)
		cmd.Dir = repoRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %v: %v\n%s", args, err, output)
		}
	}

	relOutput, err := filepath.Rel(repoRoot, output)
	if err != nil {
		t.Fatal(err)
	}
	run("run", ".", "model", "-m", mfdPath, "-o", relOutput, "-p", "db")
	run("run", ".", "repo", "-m", mfdPath, "-o", relOutput, "-p", "db", "--repo-mode", string(ModeGeneric), "-n", "portal")

	goMod := []byte(`module example.com/generated

 go 1.26

require (
	github.com/go-pg/pg/v10 v10.15.0
	github.com/go-pg/urlstruct v1.0.1
	github.com/google/uuid v1.6.0
)
`)
	if err := os.WriteFile(filepath.Join(output, "go.mod"), goMod, 0o644); err != nil {
		t.Fatal(err)
	}
	integrationTest := []byte(`//go:build integration

package db

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func TestGeneratedRuntimeCRUD(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Fatal("DB_DSN is required")
	}
	options, err := pg.ParseURL(dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := pg.Connect(options)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.Exec("TRUNCATE \"news\", \"categories\", \"tags\", \"statuses\" RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO \"statuses\" (\"statusId\") VALUES (1), (2), (3)"); err != nil {
		t.Fatal(err)
	}

	repo := NewPortalRepo(db).Category
	category, err := repo.Add(ctx, &Category{Title: "generic", OrderNumber: 1, StatusID: StatusEnabled})
	if err != nil {
		t.Fatal(err)
	}

	title := "generic"
	found, err := repo.One(ctx, (&CategorySearch{Title: &title}).Q())
	if err != nil || found == nil || found.ID != category.ID {
		t.Fatalf("One() = %#v, %v", found, err)
	}
	count, err := repo.Count(ctx, nil)
	if err != nil || count != 1 {
		t.Fatalf("Count() = %d, %v", count, err)
	}
	list, err := repo.List(ctx, nil, ApplyPager(Pager{PageSize: 10}))
	if err != nil || len(list) != 1 {
		t.Fatalf("List() = %d, %v", len(list), err)
	}

	category.Title = "updated"
	updated, err := repo.Update(ctx, category)
	if err != nil || !updated {
		t.Fatalf("Update() = %v, %v", updated, err)
	}
	deleted, err := repo.Delete(ctx, category)
	if err != nil || !deleted {
		t.Fatalf("Delete() = %v, %v", deleted, err)
	}
	count, err = repo.Count(ctx, nil)
	if err != nil || count != 0 {
		t.Fatalf("Count() after soft delete = %d, %v", count, err)
	}

	missingTitle := "missing"
	missing, err := repo.One(ctx, (&CategorySearch{Title: &missingTitle}).Q())
	if err != nil || missing != nil {
		t.Fatalf("One() for missing row = %#v, %v", missing, err)
	}

	duplicate := &Category{Title: "duplicate", OrderNumber: 2, StatusID: StatusEnabled}
	if _, err := repo.Add(ctx, duplicate); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Add(ctx, &Category{Title: "duplicate", OrderNumber: 3, StatusID: StatusEnabled}); err != nil {
		t.Fatal(err)
	}
	duplicateTitle := "duplicate"
	if _, err := repo.One(ctx, (&CategorySearch{Title: &duplicateTitle}).Q()); !errors.Is(err, pg.ErrMultiRows) {
		t.Fatalf("One() for duplicate rows error = %v", err)
	}
	cte := func(query *orm.Query) (*orm.Query, error) {
		return query.With("all_categories", db.Model(&Category{})), nil
	}
	if _, err := repo.Count(ctx, nil, cte); err != nil {
		t.Fatalf("Count() with CTE = %v", err)
	}
	deletedWithCTE, err := repo.Delete(ctx, duplicate, cte)
	if err != nil || !deletedWithCTE {
		t.Fatalf("Delete() with CTE = %v, %v", deletedWithCTE, err)
	}

	rollback := errors.New("rollback")
	err = db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		txRepo := NewPortalRepo(db).WithTransaction(tx).Category
		_, err := txRepo.Add(ctx, &Category{Title: "rolled back", OrderNumber: 4, StatusID: StatusEnabled})
		if err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("RunInTransaction() error = %v", err)
	}
	rolledBack := "rolled back"
	count, err = repo.Count(ctx, (&CategorySearch{Title: &rolledBack}).Q())
	if err != nil || count != 0 {
		t.Fatalf("Count() after rollback = %d, %v", count, err)
	}
}
`)
	if err := os.WriteFile(filepath.Join(output, "generic_runtime_test.go"), integrationTest, 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"test", "-mod=mod", "-tags=integration", "."}
	cmd := exec.Command("go", args...)
	cmd.Dir = output
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile generic output: %v\n%s", err, output)
	}
}
