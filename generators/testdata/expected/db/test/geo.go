//nolint:dupl,funlen
package test

import (
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type CityOpFunc func(t *testing.T, dbo orm.DB, in *db.City) Cleaner

func City(t *testing.T, dbo orm.DB, in *db.City, ops ...CityOpFunc) (*db.City, Cleaner) {
	repo := db.NewGeoRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.City{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		city, err := repo.CityByID(t.Context(), in.ID, repo.FullCity())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if city == nil {
			t.Fatalf("the entity City is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return city, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	city, err := repo.AddCity(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return city, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.City{ID: city.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithCityRelations(t *testing.T, dbo orm.DB, in *db.City) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Country == nil {
		in.Country = &db.Country{}
	}

	if in.Region == nil {
		in.Region = &db.Region{}
	}

	// Check if all FKs are provided. Fill them into the main struct rels

	if in.RegionID != 0 {
		in.Region.ID = in.RegionID
	}

	if in.CountryID != 0 {
		in.Country.ID = in.CountryID
	}

	// Inject relation IDs into relations which have the same relations
	in.Region.CountryID = in.CountryID
	in.Region.Country = in.Country
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Region(t, dbo, in.Region, WithRegionRelations, WithFakeRegion)
		in.Region = rel
		in.RegionID = rel.ID
		// Fill the same relations as in Region
		in.Country = rel.Country

		cleaners = append(cleaners, relatedCleaner)
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Country(t, dbo, in.Country, WithFakeCountry)
		in.Country = rel
		in.CountryID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeCity(t *testing.T, dbo orm.DB, in *db.City) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(cutS(gofakeit.Sentence(10), 255), " ", "-")
	}

	if in.OrderNumber == 0 {
		in.OrderNumber = gofakeit.IntRange(1, 10)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type CountryOpFunc func(t *testing.T, dbo orm.DB, in *db.Country) Cleaner

func Country(t *testing.T, dbo orm.DB, in *db.Country, ops ...CountryOpFunc) (*db.Country, Cleaner) {
	repo := db.NewGeoRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Country{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		country, err := repo.CountryByID(t.Context(), in.ID, repo.FullCountry())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if country == nil {
			t.Fatalf("the entity Country is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return country, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	country, err := repo.AddCountry(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return country, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Country{ID: country.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeCountry(t *testing.T, dbo orm.DB, in *db.Country) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(cutS(gofakeit.Sentence(10), 255), " ", "-")
	}

	if in.OrderNumber == 0 {
		in.OrderNumber = gofakeit.IntRange(1, 10)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type RegionOpFunc func(t *testing.T, dbo orm.DB, in *db.Region) Cleaner

func Region(t *testing.T, dbo orm.DB, in *db.Region, ops ...RegionOpFunc) (*db.Region, Cleaner) {
	repo := db.NewGeoRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Region{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		region, err := repo.RegionByID(t.Context(), in.ID, repo.FullRegion())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if region == nil {
			t.Fatalf("the entity Region is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return region, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	region, err := repo.AddRegion(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return region, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Region{ID: region.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithRegionRelations(t *testing.T, dbo orm.DB, in *db.Region) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Country == nil {
		in.Country = &db.Country{}
	}

	// Check if all FKs are provided. Fill them into the main struct rels

	if in.CountryID != 0 {
		in.Country.ID = in.CountryID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Country(t, dbo, in.Country, WithFakeCountry)
		in.Country = rel
		in.CountryID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeRegion(t *testing.T, dbo orm.DB, in *db.Region) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(cutS(gofakeit.Sentence(10), 255), " ", "-")
	}

	if in.OrderNumber == 0 {
		in.OrderNumber = gofakeit.IntRange(1, 10)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
