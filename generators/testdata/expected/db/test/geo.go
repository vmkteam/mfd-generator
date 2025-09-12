package test

import (
	"fmt"
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
			t.Fatal(fmt.Errorf("fetch the main entity City by ID=%v, err=%w", in.ID, errNotFound))
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
	// Prepare nested relations which have the same relations

	if in.Region == nil {
		in.Region = &db.Region{}
	}

	// Inject relation IDs into relations which have the same relations
	in.Region.CountryID = in.CountryID

	// Check embedded entities by FK

	// Region. Check if all FKs are provided.
	if in.RegionID != 0 {
		in.Region.ID = in.RegionID // Fill them for the next fetching step
	}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Region(t, dbo, in.Region, WithRegionRelations, WithFakeRegion)
		in.Region = rel
		// Fill the same relations as in Region
		in.Country = rel.Country

		cleaners = append(cleaners, relatedCleaner)
	}

	// Country. Check if all FKs are provided.
	if in.CountryID != 0 {
		in.Country.ID = in.CountryID // Fill them for the next fetching step
	}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Country(t, dbo, in.Country, WithFakeCountry)
		in.Country = rel

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
		in.Title = string([]rune(gofakeit.Sentence(10))[:256])
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(string([]rune(gofakeit.Sentence(10))[:256]), " ", "-")
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
			t.Fatal(fmt.Errorf("fetch the main entity Country by ID=%v, err=%w", in.ID, errNotFound))
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
		in.Title = string([]rune(gofakeit.Sentence(10))[:256])
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(string([]rune(gofakeit.Sentence(10))[:256]), " ", "-")
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
			t.Fatal(fmt.Errorf("fetch the main entity Region by ID=%v, err=%w", in.ID, errNotFound))
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

	// Check embedded entities by FK

	// Country. Check if all FKs are provided.
	if in.CountryID != 0 {
		in.Country.ID = in.CountryID // Fill them for the next fetching step
	}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Country(t, dbo, in.Country, WithFakeCountry)
		in.Country = rel

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
		in.Title = string([]rune(gofakeit.Sentence(10))[:256])
	}

	if in.Alias == "" {
		in.Alias = strings.ReplaceAll(string([]rune(gofakeit.Sentence(10))[:256]), " ", "-")
	}

	if in.OrderNumber == 0 {
		in.OrderNumber = gofakeit.IntRange(1, 10)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
