package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type CategoryOpFunc func(t *testing.T, dbo orm.DB, in *db.Category) Cleaner

func Category(t *testing.T, dbo orm.DB, in *db.Category, ops ...CategoryOpFunc) (*db.Category, Cleaner) {
	repo := db.NewPortalRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Category{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		category, err := repo.CategoryByID(t.Context(), in.ID, repo.FullCategory())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if category == nil {
			t.Fatal(fmt.Errorf("fetch the main entity Category by ID=%v, err=%w", in.ID, errNotFound))
		}

		// Return if found without real cleanup
		return category, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	category, err := repo.AddCategory(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return category, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Category{ID: category.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeCategory(t *testing.T, dbo orm.DB, in *db.Category) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.OrderNumber == 0 {
		in.OrderNumber = gofakeit.IntRange(1, 10)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type NewsOpFunc func(t *testing.T, dbo orm.DB, in *db.News) Cleaner

func News(t *testing.T, dbo orm.DB, in *db.News, ops ...NewsOpFunc) (*db.News, Cleaner) {
	repo := db.NewPortalRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.News{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		news, err := repo.NewsByID(t.Context(), in.ID, repo.FullNews())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if news == nil {
			t.Fatal(fmt.Errorf("fetch the main entity News by ID=%v, err=%w", in.ID, errNotFound))
		}

		// Return if found without real cleanup
		return news, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	news, err := repo.AddNews(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return news, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.News{ID: news.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithNewsRelations(t *testing.T, dbo orm.DB, in *db.News) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Category == nil {
		in.Category = &db.Category{}
	}

	if in.City == nil {
		in.City = &db.City{}
	}

	if in.Country == nil {
		in.Country = &db.Country{}
	}

	if in.Region == nil {
		in.Region = &db.Region{}
	}

	// Prepare nested relations which have the same relations
	if in.City.Region == nil {
		in.City.Region = &db.Region{}
	}

	// Inject relation IDs into relations which have the same relations
	in.City.Region.CountryID = val(in.CountryID)
	in.City.RegionID = val(in.RegionID)
	in.City.CountryID = val(in.CountryID)
	in.Region.CountryID = val(in.CountryID)

	// Check embedded entities by FK

	// City. Check if all FKs are provided.

	if in.CityID != nil && *in.CityID != 0 {
		in.City.ID = val(in.CityID)
	}

	if in.RegionID != nil && *in.RegionID != 0 {
		in.Region.ID = val(in.RegionID)
	}

	if in.CategoryID != 0 {
		in.Category.ID = in.CategoryID
	}

	if in.CountryID != nil && *in.CountryID != 0 {
		in.Country.ID = val(in.CountryID)
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := City(t, dbo, in.City, WithCityRelations, WithFakeCity)
		in.City = rel
		in.CityID = &rel.ID
		// Fill the same relations as in City
		in.City.Country = rel.Region.Country
		in.Region.Country = rel.Region.Country
		in.Country = rel.Region.Country
		in.Region = rel.Region

		cleaners = append(cleaners, relatedCleaner)
	}

	// Region. Check if all FKs are provided.

	if in.CityID != nil && *in.CityID != 0 {
		in.City.ID = val(in.CityID)
	}

	if in.RegionID != nil && *in.RegionID != 0 {
		in.Region.ID = val(in.RegionID)
	}

	if in.CategoryID != 0 {
		in.Category.ID = in.CategoryID
	}

	if in.CountryID != nil && *in.CountryID != 0 {
		in.Country.ID = val(in.CountryID)
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Region(t, dbo, in.Region, WithRegionRelations, WithFakeRegion)
		in.Region = rel
		in.RegionID = &rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	// Category. Check if all FKs are provided.

	if in.CityID != nil && *in.CityID != 0 {
		in.City.ID = val(in.CityID)
	}

	if in.RegionID != nil && *in.RegionID != 0 {
		in.Region.ID = val(in.RegionID)
	}

	if in.CategoryID != 0 {
		in.Category.ID = in.CategoryID
	}

	if in.CountryID != nil && *in.CountryID != 0 {
		in.Country.ID = val(in.CountryID)
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Category(t, dbo, in.Category, WithFakeCategory)
		in.Category = rel
		in.CategoryID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	// Country. Check if all FKs are provided.

	if in.CityID != nil && *in.CityID != 0 {
		in.City.ID = val(in.CityID)
	}

	if in.RegionID != nil && *in.RegionID != 0 {
		in.Region.ID = val(in.RegionID)
	}

	if in.CategoryID != 0 {
		in.Category.ID = in.CategoryID
	}

	if in.CountryID != nil && *in.CountryID != 0 {
		in.Country.ID = val(in.CountryID)
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Country(t, dbo, in.Country, WithFakeCountry)
		in.Country = rel
		in.CountryID = &rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeNews(t *testing.T, dbo orm.DB, in *db.News) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.CategoryID == 0 {
		in.CategoryID = gofakeit.IntRange(1, 10)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type TagOpFunc func(t *testing.T, dbo orm.DB, in *db.Tag) Cleaner

func Tag(t *testing.T, dbo orm.DB, in *db.Tag, ops ...TagOpFunc) (*db.Tag, Cleaner) {
	repo := db.NewPortalRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Tag{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		tag, err := repo.TagByID(t.Context(), in.ID, repo.FullTag())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if tag == nil {
			t.Fatal(fmt.Errorf("fetch the main entity Tag by ID=%v, err=%w", in.ID, errNotFound))
		}

		// Return if found without real cleanup
		return tag, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	tag, err := repo.AddTag(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return tag, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Tag{ID: tag.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeTag(t *testing.T, dbo orm.DB, in *db.Tag) Cleaner {
	if in.Title == "" {
		in.Title = cutS(gofakeit.Sentence(10), 255)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
