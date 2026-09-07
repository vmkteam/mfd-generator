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
	dbTestOutput := filepath.Join(output, "test")
	relDBTestOutput, err := filepath.Rel(repoRoot, dbTestOutput)
	if err != nil {
		t.Fatal(err)
	}
	run("run", ".", "dbtest", "-m", mfdPath, "-o", relDBTestOutput, "-p", "test", "-x", "example.com/generated/db", "--repo-mode", string(ModeGeneric), "-n", "portal")

	goMod := []byte(`module example.com/generated/db

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
	assertSentinel := func(name string, invoke func(Applier) error) {
		t.Helper()
		sentinel := errors.New(name + " applier")
		applier := func(query *orm.Query) (*orm.Query, error) {
			return query, sentinel
		}
		if err := invoke(applier); !errors.Is(err, sentinel) {
			t.Fatalf("%s() sentinel applier error = %v", name, err)
		}
	}
	if _, err := db.Exec("TRUNCATE \"news\", \"categories\", \"tags\", \"statuses\" RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO \"statuses\" (\"statusId\") VALUES (1), (2), (3)"); err != nil {
		t.Fatal(err)
	}

	repo := NewPortalRepo(db).Category
	assertSentinel("Add", func(applier Applier) error {
		_, err := repo.Add(ctx, &Category{Title: "sentinel", OrderNumber: 1, StatusID: StatusEnabled}, applier)
		return err
	})
	category, err := repo.Add(ctx, &Category{Title: "generic", OrderNumber: 1, StatusID: StatusEnabled})
	if err != nil {
		t.Fatal(err)
	}

	title := "generic"
	found, err := repo.One(ctx, (&CategorySearch{Title: &title}).Q())
	if err != nil || found == nil || found.ID != category.ID {
		t.Fatalf("One() = %#v, %v", found, err)
	}
	assertSentinel("One", func(applier Applier) error {
		_, err := repo.One(ctx, nil, applier)
		return err
	})
	assertSentinel("applyAppliers", func(applier Applier) error {
		_, err := applyAppliers(repo.db.ModelContext(ctx, &Category{}), applier)
		return err
	})
	searchApplierErr := errors.New("search applier")
	search := &CategorySearch{}
	search.WithApply(func(query *orm.Query) (*orm.Query, error) {
		return query, searchApplierErr
	})
	if _, err := search.Q()(repo.db.ModelContext(ctx, &Category{})); !errors.Is(err, searchApplierErr) {
		t.Fatalf("Q() search applier error = %v", err)
	}
	if _, err := repo.One(ctx, search.Q()); !errors.Is(err, searchApplierErr) {
		t.Fatalf("One() search applier error = %v", err)
	}
	assertSentinel("List", func(applier Applier) error {
		_, err := repo.List(ctx, nil, nil, applier)
		return err
	})
	assertSentinel("Count", func(applier Applier) error {
		_, err := repo.Count(ctx, nil, applier)
		return err
	})
	count, err := repo.Count(ctx, nil)
	if err != nil || count != 1 {
		t.Fatalf("Count() = %d, %v", count, err)
	}
	list, err := repo.List(ctx, nil, ApplyPager(Pager{PageSize: 10}))
	if err != nil || len(list) != 1 {
		t.Fatalf("List() = %d, %v", len(list), err)
	}

	assertSentinel("Update", func(applier Applier) error {
		_, err := repo.Update(ctx, category, applier)
		return err
	})
	category.Title = "updated"
	updated, err := repo.Update(ctx, category)
	if err != nil || !updated {
		t.Fatalf("Update() = %v, %v", updated, err)
	}
	assertSentinel("Delete", func(applier Applier) error {
		_, err := repo.Delete(ctx, &Category{ID: category.ID, StatusID: StatusEnabled}, applier)
		return err
	})
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
	cteDefinition := func(query *orm.Query) (*orm.Query, error) {
		return query.With("all_categories", db.Model(&Category{}).Where("\"categoryId\" = ?", duplicate.ID)), nil
	}
	ctePredicate := func(query *orm.Query) (*orm.Query, error) {
		return query.Where("EXISTS (SELECT 1 FROM \"all_categories\" WHERE \"all_categories\".\"categoryId\" = ?)", duplicate.ID), nil
	}
	if _, err := repo.Count(ctx, nil, cteDefinition, ctePredicate); err != nil {
		t.Fatalf("Count() with CTE = %v", err)
	}
	staleDuplicate := &Category{
		ID:          duplicate.ID,
		Title:       "stale title",
		OrderNumber: 999,
		StatusID:    StatusEnabled,
	}
	explicitDeleteColumn := ApplyOp(WithColumns(Columns.Category.Title))
	falsePredicate := ApplyOp(func(query *orm.Query) {
		query.Where("1 = 0")
	})
	deletedWithFalsePredicate, err := repo.Delete(ctx, staleDuplicate, cteDefinition, ctePredicate, falsePredicate, explicitDeleteColumn)
	if err != nil || deletedWithFalsePredicate {
		t.Fatalf("Delete() with false predicate = %v, %v", deletedWithFalsePredicate, err)
	}
	stillActive := &Category{ID: duplicate.ID}
	if err := db.Model(stillActive).WherePK().Select(); err != nil {
		t.Fatal(err)
	}
	if stillActive.StatusID != StatusEnabled || stillActive.Title != "duplicate" || stillActive.OrderNumber != 2 {
		t.Fatalf("Delete() with false predicate changed target: %#v", stillActive)
	}

	deletedWithCTE, err := repo.Delete(ctx, staleDuplicate, cteDefinition, ctePredicate, explicitDeleteColumn)
	if err != nil || !deletedWithCTE {
		t.Fatalf("Delete() with CTE = %v, %v", deletedWithCTE, err)
	}
	storedDuplicate := &Category{ID: duplicate.ID}
	if err := db.Model(storedDuplicate).WherePK().Select(); err != nil {
		t.Fatal(err)
	}
	if storedDuplicate.StatusID != StatusDeleted {
		t.Fatalf("Delete() with CTE statusId = %d, want %d", storedDuplicate.StatusID, StatusDeleted)
	}
	if storedDuplicate.Title != "duplicate" || storedDuplicate.OrderNumber != 2 {
		t.Fatalf("Delete() with CTE overwrote columns: %#v", storedDuplicate)
	}
	duplicateCount, err := repo.Count(ctx, (&CategorySearch{Title: &duplicateTitle}).Q())
	if err != nil || duplicateCount != 1 {
		t.Fatalf("Count() after soft delete with CTE = %d, %v", duplicateCount, err)
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

	args := []string{"test", "-mod=mod", "-tags=integration", "./..."}
	cmd := exec.Command("go", args...)
	cmd.Dir = output
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile generic output: %v\n%s", err, output)
	}
}
