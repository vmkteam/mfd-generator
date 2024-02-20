package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type PortalRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewPortalRepo returns new repository
func NewPortalRepo(db orm.DB) PortalRepo {
	return PortalRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.Category.Name: {StatusFilter},
			Tables.News.Name:     {StatusFilter},
			Tables.Tag.Name:      {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.Category.Name: {{Column: Columns.Category.Title, Direction: SortAsc}},
			Tables.News.Name:     {{Column: Columns.News.CreatedAt, Direction: SortDesc}},
			Tables.Tag.Name:      {{Column: Columns.Tag.Title, Direction: SortAsc}},
		},
		join: map[string][]string{
			Tables.Category.Name: {TableColumns},
			Tables.News.Name:     {TableColumns, Columns.News.Category},
			Tables.Tag.Name:      {TableColumns},
		},
	}
}

// WithTransaction is a function that wraps PortalRepo with pg.Tx transaction.
func (pr PortalRepo) WithTransaction(tx *pg.Tx) PortalRepo {
	pr.db = tx
	return pr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (pr PortalRepo) WithEnabledOnly() PortalRepo {
	f := make(map[string][]Filter, len(pr.filters))
	for i := range pr.filters {
		f[i] = make([]Filter, len(pr.filters[i]))
		copy(f[i], pr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	pr.filters = f

	return pr
}

/*** Category ***/

// FullCategory returns full joins with all columns
func (pr PortalRepo) FullCategory() OpFunc {
	return WithColumns(pr.join[Tables.Category.Name]...)
}

// DefaultCategorySort returns default sort.
func (pr PortalRepo) DefaultCategorySort() OpFunc {
	return WithSort(pr.sort[Tables.Category.Name]...)
}

// CategoryByID is a function that returns Category by ID(s) or nil.
func (pr PortalRepo) CategoryByID(ctx context.Context, id int, ops ...OpFunc) (*Category, error) {
	return pr.OneCategory(ctx, &CategorySearch{ID: &id}, ops...)
}

// OneCategory is a function that returns one Category by filters. It could return pg.ErrMultiRows.
func (pr PortalRepo) OneCategory(ctx context.Context, search *CategorySearch, ops ...OpFunc) (*Category, error) {
	obj := &Category{}
	err := buildQuery(ctx, pr.db, obj, search, pr.filters[Tables.Category.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// CategoriesByFilters returns Category list.
func (pr PortalRepo) CategoriesByFilters(ctx context.Context, search *CategorySearch, pager Pager, ops ...OpFunc) (categories []Category, err error) {
	err = buildQuery(ctx, pr.db, &categories, search, pr.filters[Tables.Category.Name], pager, ops...).Select()
	return
}

// CountCategories returns count
func (pr PortalRepo) CountCategories(ctx context.Context, search *CategorySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, pr.db, &Category{}, search, pr.filters[Tables.Category.Name], PagerOne, ops...).Count()
}

// AddCategory adds Category to DB.
func (pr PortalRepo) AddCategory(ctx context.Context, category *Category, ops ...OpFunc) (*Category, error) {
	q := pr.db.ModelContext(ctx, category)
	applyOps(q, ops...)
	_, err := q.Insert()

	return category, err
}

// UpdateCategory updates Category in DB.
func (pr PortalRepo) UpdateCategory(ctx context.Context, category *Category, ops ...OpFunc) (bool, error) {
	q := pr.db.ModelContext(ctx, category).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Category.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteCategory set statusId to deleted in DB.
func (pr PortalRepo) DeleteCategory(ctx context.Context, id int) (deleted bool, err error) {
	category := &Category{ID: id, StatusID: StatusDeleted}

	return pr.UpdateCategory(ctx, category, WithColumns(Columns.Category.StatusID))
}

/*** News ***/

// FullNews returns full joins with all columns
func (pr PortalRepo) FullNews() OpFunc {
	return WithColumns(pr.join[Tables.News.Name]...)
}

// DefaultNewsSort returns default sort.
func (pr PortalRepo) DefaultNewsSort() OpFunc {
	return WithSort(pr.sort[Tables.News.Name]...)
}

// NewsByID is a function that returns News by ID(s) or nil.
func (pr PortalRepo) NewsByID(ctx context.Context, id int, ops ...OpFunc) (*News, error) {
	return pr.OneNews(ctx, &NewsSearch{ID: &id}, ops...)
}

// OneNews is a function that returns one News by filters. It could return pg.ErrMultiRows.
func (pr PortalRepo) OneNews(ctx context.Context, search *NewsSearch, ops ...OpFunc) (*News, error) {
	obj := &News{}
	err := buildQuery(ctx, pr.db, obj, search, pr.filters[Tables.News.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// NewsByFilters returns News list.
func (pr PortalRepo) NewsByFilters(ctx context.Context, search *NewsSearch, pager Pager, ops ...OpFunc) (newsList []News, err error) {
	err = buildQuery(ctx, pr.db, &newsList, search, pr.filters[Tables.News.Name], pager, ops...).Select()
	return
}

// CountNews returns count
func (pr PortalRepo) CountNews(ctx context.Context, search *NewsSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, pr.db, &News{}, search, pr.filters[Tables.News.Name], PagerOne, ops...).Count()
}

// AddNews adds News to DB.
func (pr PortalRepo) AddNews(ctx context.Context, news *News, ops ...OpFunc) (*News, error) {
	q := pr.db.ModelContext(ctx, news)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.News.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return news, err
}

// UpdateNews updates News in DB.
func (pr PortalRepo) UpdateNews(ctx context.Context, news *News, ops ...OpFunc) (bool, error) {
	q := pr.db.ModelContext(ctx, news).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.News.ID, Columns.News.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteNews set statusId to deleted in DB.
func (pr PortalRepo) DeleteNews(ctx context.Context, id int) (deleted bool, err error) {
	news := &News{ID: id, StatusID: StatusDeleted}

	return pr.UpdateNews(ctx, news, WithColumns(Columns.News.StatusID))
}

/*** Tag ***/

// FullTag returns full joins with all columns
func (pr PortalRepo) FullTag() OpFunc {
	return WithColumns(pr.join[Tables.Tag.Name]...)
}

// DefaultTagSort returns default sort.
func (pr PortalRepo) DefaultTagSort() OpFunc {
	return WithSort(pr.sort[Tables.Tag.Name]...)
}

// TagByID is a function that returns Tag by ID(s) or nil.
func (pr PortalRepo) TagByID(ctx context.Context, id int, ops ...OpFunc) (*Tag, error) {
	return pr.OneTag(ctx, &TagSearch{ID: &id}, ops...)
}

// OneTag is a function that returns one Tag by filters. It could return pg.ErrMultiRows.
func (pr PortalRepo) OneTag(ctx context.Context, search *TagSearch, ops ...OpFunc) (*Tag, error) {
	obj := &Tag{}
	err := buildQuery(ctx, pr.db, obj, search, pr.filters[Tables.Tag.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TagsByFilters returns Tag list.
func (pr PortalRepo) TagsByFilters(ctx context.Context, search *TagSearch, pager Pager, ops ...OpFunc) (tags []Tag, err error) {
	err = buildQuery(ctx, pr.db, &tags, search, pr.filters[Tables.Tag.Name], pager, ops...).Select()
	return
}

// CountTags returns count
func (pr PortalRepo) CountTags(ctx context.Context, search *TagSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, pr.db, &Tag{}, search, pr.filters[Tables.Tag.Name], PagerOne, ops...).Count()
}

// AddTag adds Tag to DB.
func (pr PortalRepo) AddTag(ctx context.Context, tag *Tag, ops ...OpFunc) (*Tag, error) {
	q := pr.db.ModelContext(ctx, tag)
	applyOps(q, ops...)
	_, err := q.Insert()

	return tag, err
}

// UpdateTag updates Tag in DB.
func (pr PortalRepo) UpdateTag(ctx context.Context, tag *Tag, ops ...OpFunc) (bool, error) {
	q := pr.db.ModelContext(ctx, tag).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Tag.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTag set statusId to deleted in DB.
func (pr PortalRepo) DeleteTag(ctx context.Context, id int) (deleted bool, err error) {
	tag := &Tag{ID: id, StatusID: StatusDeleted}

	return pr.UpdateTag(ctx, tag, WithColumns(Columns.Tag.StatusID))
}
