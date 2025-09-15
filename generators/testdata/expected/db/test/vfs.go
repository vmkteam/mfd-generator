package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type VfsFileOpFunc func(t *testing.T, dbo orm.DB, in *db.VfsFile) Cleaner

func VfsFile(t *testing.T, dbo orm.DB, in *db.VfsFile, ops ...VfsFileOpFunc) (*db.VfsFile, Cleaner) {
	repo := db.NewVfsRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.VfsFile{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		vfsFile, err := repo.VfsFileByID(t.Context(), in.ID, repo.FullVfsFile())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if vfsFile == nil {
			t.Fatal(fmt.Errorf("fetch the main entity VfsFile by ID=%v, err=%w", in.ID, errNotFound))
		}

		// Return if found without real cleanup
		return vfsFile, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	vfsFile, err := repo.AddVfsFile(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return vfsFile, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.VfsFile{ID: vfsFile.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithVfsFileRelations(t *testing.T, dbo orm.DB, in *db.VfsFile) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Folder == nil {
		in.Folder = &db.VfsFolder{}
	}

	// Check embedded entities by FK

	// Folder. Check if all FKs are provided.

	if in.FolderID == 0 {
		in.Folder.ID = in.FolderID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := VfsFolder(t, dbo, in.Folder, WithFakeVfsFolder)
		in.Folder = rel
		in.FolderID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeVfsFile(t *testing.T, dbo orm.DB, in *db.VfsFile) Cleaner {
	if in.FolderID == 0 {
		in.FolderID = gofakeit.IntRange(1, 10)
	}

	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.Path == "" {
		in.Path = cutS(gofakeit.Sentence(10), 255)
	}

	if in.MimeType == "" {
		in.MimeType = cutS(gofakeit.Sentence(10), 255)
	}

	if in.FileExists == false {
		in.FileExists = gofakeit.Bool()
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type VfsFolderOpFunc func(t *testing.T, dbo orm.DB, in *db.VfsFolder) Cleaner

func VfsFolder(t *testing.T, dbo orm.DB, in *db.VfsFolder, ops ...VfsFolderOpFunc) (*db.VfsFolder, Cleaner) {
	repo := db.NewVfsRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.VfsFolder{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		vfsFolder, err := repo.VfsFolderByID(t.Context(), in.ID, repo.FullVfsFolder())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if vfsFolder == nil {
			t.Fatal(fmt.Errorf("fetch the main entity VfsFolder by ID=%v, err=%w", in.ID, errNotFound))
		}

		// Return if found without real cleanup
		return vfsFolder, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	vfsFolder, err := repo.AddVfsFolder(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return vfsFolder, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.VfsFolder{ID: vfsFolder.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeVfsFolder(t *testing.T, dbo orm.DB, in *db.VfsFolder) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
