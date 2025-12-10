//nolint:dupl,funlen
package test

import (
	"testing"
	"time"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type LoginCodeOpFunc func(t *testing.T, dbo orm.DB, in *db.LoginCode) Cleaner

func LoginCode(t *testing.T, dbo orm.DB, in *db.LoginCode, ops ...LoginCodeOpFunc) (*db.LoginCode, Cleaner) {
	repo := db.NewCommonRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.LoginCode{}
	}

	// Check if PKs are provided
	if in.ID != "" {
		// Fetch the entity by PK
		loginCode, err := repo.LoginCodeByID(t.Context(), in.ID, repo.FullLoginCode())
		if err != nil {
			t.Fatal(err)
		}
		// Return if found without real cleanup
		if loginCode != nil {
			return loginCode, emptyClean
		}

		// If we're here, we don't find the entity by PKs. Just try to add the entity by provided PK
		t.Logf("the entity LoginCode is not found by provided PKs:ID=%v. Trying to create one", in.ID)
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	loginCode, err := repo.AddLoginCode(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return loginCode, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.LoginCode{ID: loginCode.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithLoginCodeRelations(t *testing.T, dbo orm.DB, in *db.LoginCode) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.SiteUser == nil {
		in.SiteUser = &db.SiteUser{}
	}

	// Check if all FKs are provided. Fill them into the main struct rels

	if in.SiteUserID != 0 {
		in.SiteUser.ID = in.SiteUserID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := SiteUser(t, dbo, in.SiteUser, WithFakeSiteUser)
		in.SiteUser = rel
		in.SiteUserID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeLoginCode(t *testing.T, dbo orm.DB, in *db.LoginCode) Cleaner {
	if in.Code == "" {
		in.Code = cutS(gofakeit.Word(), 8)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.Attempts == 0 {
		in.Attempts = gofakeit.IntRange(1, 10)
	}

	return emptyClean
}

type SiteUserOpFunc func(t *testing.T, dbo orm.DB, in *db.SiteUser) Cleaner

func SiteUser(t *testing.T, dbo orm.DB, in *db.SiteUser, ops ...SiteUserOpFunc) (*db.SiteUser, Cleaner) {
	repo := db.NewCommonRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.SiteUser{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		siteUser, err := repo.SiteUserByID(t.Context(), in.ID, repo.FullSiteUser())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if siteUser == nil {
			t.Fatalf("the entity SiteUser is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return siteUser, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	siteUser, err := repo.AddSiteUser(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return siteUser, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.SiteUser{ID: siteUser.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeSiteUser(t *testing.T, dbo orm.DB, in *db.SiteUser) Cleaner {
	if in.StatusID == 0 {
		in.StatusID = 1
	}

	if in.Email == "" {
		in.Email = cutS(gofakeit.Email(), 255)
	}

	if in.DefaultRole == "" {
		in.DefaultRole = cutS(gofakeit.Sentence(3), 32)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	return emptyClean
}
