//nolint:dupl,funlen
package test

import (
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type EncryptionKeyOpFunc func(t *testing.T, dbo orm.DB, in *db.EncryptionKey) Cleaner

func EncryptionKey(t *testing.T, dbo orm.DB, in *db.EncryptionKey, ops ...EncryptionKeyOpFunc) (*db.EncryptionKey, Cleaner) {
	repo := db.NewCardRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.EncryptionKey{}
	}

	// Check if PKs are provided
	var defID uuid.UUID
	if in.ID != defID {
		// Fetch the entity by PK
		encryptionKey, err := repo.EncryptionKeyByID(t.Context(), in.ID, repo.FullEncryptionKey())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if encryptionKey == nil {
			t.Fatalf("the entity EncryptionKey is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return encryptionKey, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	encryptionKey, err := repo.AddEncryptionKey(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return encryptionKey, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.EncryptionKey{ID: encryptionKey.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeEncryptionKey(t *testing.T, dbo orm.DB, in *db.EncryptionKey) Cleaner {
	if in.IssuedCount == 0 {
		in.IssuedCount = gofakeit.IntRange(1, 10)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.ExpiresAt.IsZero() {
		in.ExpiresAt = gofakeit.DateRange(time.Now().Add(5*time.Minute), time.Now().Add(1*time.Hour))
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
