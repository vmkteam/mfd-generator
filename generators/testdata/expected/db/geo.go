package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type GeoRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewGeoRepo returns new repository
func NewGeoRepo(db orm.DB) GeoRepo {
	return GeoRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.City.Name:    {StatusFilter},
			Tables.Country.Name: {StatusFilter},
			Tables.Region.Name:  {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.City.Name:    {{Column: Columns.City.Title, Direction: SortAsc}},
			Tables.Country.Name: {{Column: Columns.Country.Title, Direction: SortAsc}},
			Tables.Region.Name:  {{Column: Columns.Region.Title, Direction: SortAsc}},
		},
		join: map[string][]string{
			Tables.City.Name:    {TableColumns, Columns.City.Region, Columns.City.Country},
			Tables.Country.Name: {TableColumns},
			Tables.Region.Name:  {TableColumns, Columns.Region.Country},
		},
	}
}

// WithTransaction is a function that wraps GeoRepo with pg.Tx transaction.
func (gr GeoRepo) WithTransaction(tx *pg.Tx) GeoRepo {
	gr.db = tx
	return gr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (gr GeoRepo) WithEnabledOnly() GeoRepo {
	f := make(map[string][]Filter, len(gr.filters))
	for i := range gr.filters {
		f[i] = make([]Filter, len(gr.filters[i]))
		copy(f[i], gr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	gr.filters = f

	return gr
}

/*** City ***/

// FullCity returns full joins with all columns
func (gr GeoRepo) FullCity() OpFunc {
	return WithColumns(gr.join[Tables.City.Name]...)
}

// DefaultCitySort returns default sort.
func (gr GeoRepo) DefaultCitySort() OpFunc {
	return WithSort(gr.sort[Tables.City.Name]...)
}

// CityByID is a function that returns City by ID(s) or nil.
func (gr GeoRepo) CityByID(ctx context.Context, id int, ops ...OpFunc) (*City, error) {
	return gr.OneCity(ctx, &CitySearch{ID: &id}, ops...)
}

// OneCity is a function that returns one City by filters. It could return pg.ErrMultiRows.
func (gr GeoRepo) OneCity(ctx context.Context, search *CitySearch, ops ...OpFunc) (*City, error) {
	obj := &City{}
	err := buildQuery(ctx, gr.db, obj, search, gr.filters[Tables.City.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// CitiesByFilters returns City list.
func (gr GeoRepo) CitiesByFilters(ctx context.Context, search *CitySearch, pager Pager, ops ...OpFunc) (cities []City, err error) {
	err = buildQuery(ctx, gr.db, &cities, search, gr.filters[Tables.City.Name], pager, ops...).Select()
	return
}

// CountCities returns count
func (gr GeoRepo) CountCities(ctx context.Context, search *CitySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, gr.db, &City{}, search, gr.filters[Tables.City.Name], PagerOne, ops...).Count()
}

// AddCity adds City to DB.
func (gr GeoRepo) AddCity(ctx context.Context, city *City, ops ...OpFunc) (*City, error) {
	q := gr.db.ModelContext(ctx, city)
	applyOps(q, ops...)
	_, err := q.Insert()

	return city, err
}

// UpdateCity updates City in DB.
func (gr GeoRepo) UpdateCity(ctx context.Context, city *City, ops ...OpFunc) (bool, error) {
	q := gr.db.ModelContext(ctx, city).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.City.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteCity set statusId to deleted in DB.
func (gr GeoRepo) DeleteCity(ctx context.Context, id int) (deleted bool, err error) {
	city := &City{ID: id, StatusID: StatusDeleted}

	return gr.UpdateCity(ctx, city, WithColumns(Columns.City.StatusID))
}

/*** Country ***/

// FullCountry returns full joins with all columns
func (gr GeoRepo) FullCountry() OpFunc {
	return WithColumns(gr.join[Tables.Country.Name]...)
}

// DefaultCountrySort returns default sort.
func (gr GeoRepo) DefaultCountrySort() OpFunc {
	return WithSort(gr.sort[Tables.Country.Name]...)
}

// CountryByID is a function that returns Country by ID(s) or nil.
func (gr GeoRepo) CountryByID(ctx context.Context, id int, ops ...OpFunc) (*Country, error) {
	return gr.OneCountry(ctx, &CountrySearch{ID: &id}, ops...)
}

// OneCountry is a function that returns one Country by filters. It could return pg.ErrMultiRows.
func (gr GeoRepo) OneCountry(ctx context.Context, search *CountrySearch, ops ...OpFunc) (*Country, error) {
	obj := &Country{}
	err := buildQuery(ctx, gr.db, obj, search, gr.filters[Tables.Country.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// CountriesByFilters returns Country list.
func (gr GeoRepo) CountriesByFilters(ctx context.Context, search *CountrySearch, pager Pager, ops ...OpFunc) (countries []Country, err error) {
	err = buildQuery(ctx, gr.db, &countries, search, gr.filters[Tables.Country.Name], pager, ops...).Select()
	return
}

// CountCountries returns count
func (gr GeoRepo) CountCountries(ctx context.Context, search *CountrySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, gr.db, &Country{}, search, gr.filters[Tables.Country.Name], PagerOne, ops...).Count()
}

// AddCountry adds Country to DB.
func (gr GeoRepo) AddCountry(ctx context.Context, country *Country, ops ...OpFunc) (*Country, error) {
	q := gr.db.ModelContext(ctx, country)
	applyOps(q, ops...)
	_, err := q.Insert()

	return country, err
}

// UpdateCountry updates Country in DB.
func (gr GeoRepo) UpdateCountry(ctx context.Context, country *Country, ops ...OpFunc) (bool, error) {
	q := gr.db.ModelContext(ctx, country).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Country.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteCountry set statusId to deleted in DB.
func (gr GeoRepo) DeleteCountry(ctx context.Context, id int) (deleted bool, err error) {
	country := &Country{ID: id, StatusID: StatusDeleted}

	return gr.UpdateCountry(ctx, country, WithColumns(Columns.Country.StatusID))
}

/*** Region ***/

// FullRegion returns full joins with all columns
func (gr GeoRepo) FullRegion() OpFunc {
	return WithColumns(gr.join[Tables.Region.Name]...)
}

// DefaultRegionSort returns default sort.
func (gr GeoRepo) DefaultRegionSort() OpFunc {
	return WithSort(gr.sort[Tables.Region.Name]...)
}

// RegionByID is a function that returns Region by ID(s) or nil.
func (gr GeoRepo) RegionByID(ctx context.Context, id int, ops ...OpFunc) (*Region, error) {
	return gr.OneRegion(ctx, &RegionSearch{ID: &id}, ops...)
}

// OneRegion is a function that returns one Region by filters. It could return pg.ErrMultiRows.
func (gr GeoRepo) OneRegion(ctx context.Context, search *RegionSearch, ops ...OpFunc) (*Region, error) {
	obj := &Region{}
	err := buildQuery(ctx, gr.db, obj, search, gr.filters[Tables.Region.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// RegionsByFilters returns Region list.
func (gr GeoRepo) RegionsByFilters(ctx context.Context, search *RegionSearch, pager Pager, ops ...OpFunc) (regions []Region, err error) {
	err = buildQuery(ctx, gr.db, &regions, search, gr.filters[Tables.Region.Name], pager, ops...).Select()
	return
}

// CountRegions returns count
func (gr GeoRepo) CountRegions(ctx context.Context, search *RegionSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, gr.db, &Region{}, search, gr.filters[Tables.Region.Name], PagerOne, ops...).Count()
}

// AddRegion adds Region to DB.
func (gr GeoRepo) AddRegion(ctx context.Context, region *Region, ops ...OpFunc) (*Region, error) {
	q := gr.db.ModelContext(ctx, region)
	applyOps(q, ops...)
	_, err := q.Insert()

	return region, err
}

// UpdateRegion updates Region in DB.
func (gr GeoRepo) UpdateRegion(ctx context.Context, region *Region, ops ...OpFunc) (bool, error) {
	q := gr.db.ModelContext(ctx, region).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Region.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteRegion set statusId to deleted in DB.
func (gr GeoRepo) DeleteRegion(ctx context.Context, id int) (deleted bool, err error) {
	region := &Region{ID: id, StatusID: StatusDeleted}

	return gr.UpdateRegion(ctx, region, WithColumns(Columns.Region.StatusID))
}
