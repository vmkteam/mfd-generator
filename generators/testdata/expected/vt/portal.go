package vt

import (
	"context"

	"github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
	"github.com/vmkteam/mfd-generator/generators/testdata/expected/embedlog"

	"github.com/vmkteam/zenrpc/v2"
)

type CategoryService struct {
	zenrpc.Service
	embedlog.Logger
	portalRepo db.PortalRepo
}

func NewCategoryService(dbo db.DB, logger embedlog.Logger) *CategoryService {
	return &CategoryService{
		Logger:     logger,
		portalRepo: db.NewPortalRepo(dbo),
	}
}

func (s CategoryService) dbSort(ops *ViewOps) db.OpFunc {
	v := s.portalRepo.DefaultCategorySort()
	if ops == nil {
		return v
	}

	switch ops.SortColumn {
	case db.Columns.Category.ID, db.Columns.Category.Title, db.Columns.Category.OrderNumber, db.Columns.Category.StatusID:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}

	return v
}

// Count returns count Categories according to conditions in search params.
//
//zenrpc:search CategorySearch
//zenrpc:return int
//zenrpc:500 Internal Error
func (s CategoryService) Count(ctx context.Context, search *CategorySearch) (int, error) {
	count, err := s.portalRepo.CountCategories(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get returns а list of Categories according to conditions in search params.
//
//zenrpc:search CategorySearch
//zenrpc:viewOps ViewOps
//zenrpc:return []CategorySummary
//zenrpc:500 Internal Error
func (s CategoryService) Get(ctx context.Context, search *CategorySearch, viewOps *ViewOps) ([]CategorySummary, error) {
	list, err := s.portalRepo.CategoriesByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.portalRepo.FullCategory())
	if err != nil {
		return nil, InternalError(err)
	}
	categories := make([]CategorySummary, 0, len(list))
	for i := 0; i < len(list); i++ {
		if category := NewCategorySummary(&list[i]); category != nil {
			categories = append(categories, *category)
		}
	}
	return categories, nil
}

// GetByID returns a Category by its ID.
//
//zenrpc:id int
//zenrpc:return Category
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s CategoryService) GetByID(ctx context.Context, id int) (*Category, error) {
	db, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewCategory(db), nil
}

func (s CategoryService) byID(ctx context.Context, id int) (*db.Category, error) {
	db, err := s.portalRepo.CategoryByID(ctx, id, s.portalRepo.FullCategory())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}

// Add adds a Category from the query.
//
//zenrpc:category Category
//zenrpc:return Category
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s CategoryService) Add(ctx context.Context, category Category) (*Category, error) {
	if ve := s.isValid(ctx, category, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	db, err := s.portalRepo.AddCategory(ctx, category.ToDB())
	if err != nil {
		return nil, InternalError(err)
	}
	return NewCategory(db), nil
}

// Update updates the Category data identified by id from the query.
//
//zenrpc:categories Category
//zenrpc:return Category
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s CategoryService) Update(ctx context.Context, category Category) (bool, error) {
	if _, err := s.byID(ctx, category.ID); err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, category, true); ve.HasErrors() {
		return false, ve.Error()
	}

	ok, err := s.portalRepo.UpdateCategory(ctx, category.ToDB())
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete deletes the Category by its ID.
//
//zenrpc:id int
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s CategoryService) Delete(ctx context.Context, id int) (bool, error) {
	if _, err := s.byID(ctx, id); err != nil {
		return false, err
	}

	ok, err := s.portalRepo.DeleteCategory(ctx, id)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Validate verifies that Category data is valid.
//
//zenrpc:category Category
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s CategoryService) Validate(ctx context.Context, category Category) ([]FieldError, error) {
	isUpdate := category.ID != 0
	if isUpdate {
		_, err := s.byID(ctx, category.ID)
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, category, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s CategoryService) isValid(ctx context.Context, category Category, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, category); v.HasInternalError() {
		return v
	}

	//custom validation starts here
	return v
}

type NewsService struct {
	zenrpc.Service
	embedlog.Logger
	portalRepo db.PortalRepo
}

func NewNewsService(dbo db.DB, logger embedlog.Logger) *NewsService {
	return &NewsService{
		Logger:     logger,
		portalRepo: db.NewPortalRepo(dbo),
	}
}

func (s NewsService) dbSort(ops *ViewOps) db.OpFunc {
	v := s.portalRepo.DefaultNewsSort()
	if ops == nil {
		return v
	}

	switch ops.SortColumn {
	case db.Columns.News.ID, db.Columns.News.Title, db.Columns.News.Preview, db.Columns.News.Content, db.Columns.News.CategoryID, db.Columns.News.CreatedAt, db.Columns.News.PublishedAt, db.Columns.News.StatusID:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}

	return v
}

// Count returns count News according to conditions in search params.
//
//zenrpc:search NewsSearch
//zenrpc:return int
//zenrpc:500 Internal Error
func (s NewsService) Count(ctx context.Context, search *NewsSearch) (int, error) {
	count, err := s.portalRepo.CountNews(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get returns а list of News according to conditions in search params.
//
//zenrpc:search NewsSearch
//zenrpc:viewOps ViewOps
//zenrpc:return []NewsSummary
//zenrpc:500 Internal Error
func (s NewsService) Get(ctx context.Context, search *NewsSearch, viewOps *ViewOps) ([]NewsSummary, error) {
	list, err := s.portalRepo.NewsByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.portalRepo.FullNews())
	if err != nil {
		return nil, InternalError(err)
	}
	newsList := make([]NewsSummary, 0, len(list))
	for i := 0; i < len(list); i++ {
		if news := NewNewsSummary(&list[i]); news != nil {
			newsList = append(newsList, *news)
		}
	}
	return newsList, nil
}

// GetByID returns a News by its ID.
//
//zenrpc:id int
//zenrpc:return News
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s NewsService) GetByID(ctx context.Context, id int) (*News, error) {
	db, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewNews(db), nil
}

func (s NewsService) byID(ctx context.Context, id int) (*db.News, error) {
	db, err := s.portalRepo.NewsByID(ctx, id, s.portalRepo.FullNews())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}

// Add adds a News from the query.
//
//zenrpc:news News
//zenrpc:return News
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s NewsService) Add(ctx context.Context, news News) (*News, error) {
	if ve := s.isValid(ctx, news, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	db, err := s.portalRepo.AddNews(ctx, news.ToDB())
	if err != nil {
		return nil, InternalError(err)
	}
	return NewNews(db), nil
}

// Update updates the News data identified by id from the query.
//
//zenrpc:newsList News
//zenrpc:return News
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s NewsService) Update(ctx context.Context, news News) (bool, error) {
	if _, err := s.byID(ctx, news.ID); err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, news, true); ve.HasErrors() {
		return false, ve.Error()
	}

	ok, err := s.portalRepo.UpdateNews(ctx, news.ToDB())
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete deletes the News by its ID.
//
//zenrpc:id int
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s NewsService) Delete(ctx context.Context, id int) (bool, error) {
	if _, err := s.byID(ctx, id); err != nil {
		return false, err
	}

	ok, err := s.portalRepo.DeleteNews(ctx, id)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Validate verifies that News data is valid.
//
//zenrpc:news News
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s NewsService) Validate(ctx context.Context, news News) ([]FieldError, error) {
	isUpdate := news.ID != 0
	if isUpdate {
		_, err := s.byID(ctx, news.ID)
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, news, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s NewsService) isValid(ctx context.Context, news News, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, news); v.HasInternalError() {
		return v
	}

	// check fks
	if news.CategoryID != 0 {
		item, err := s.portalRepo.CategoryByID(ctx, news.CategoryID)
		if err != nil {
			v.SetInternalError(err)
		} else if item == nil {
			v.Append("categoryId", FieldErrorIncorrect)
		}
	}

	if len(news.TagIDs) != 0 {
		items, err := s.portalRepo.TagsByFilters(ctx, &db.TagSearch{IDs: news.TagIDs}, db.PagerNoLimit)
		if err != nil {
			v.SetInternalError(err)
		} else if len(items) != len(news.TagIDs) {
			v.Append("tagIds", FieldErrorIncorrect)
		}
	}
	//custom validation starts here
	return v
}

type TagService struct {
	zenrpc.Service
	embedlog.Logger
	portalRepo db.PortalRepo
}

func NewTagService(dbo db.DB, logger embedlog.Logger) *TagService {
	return &TagService{
		Logger:     logger,
		portalRepo: db.NewPortalRepo(dbo),
	}
}

func (s TagService) dbSort(ops *ViewOps) db.OpFunc {
	v := s.portalRepo.DefaultTagSort()
	if ops == nil {
		return v
	}

	switch ops.SortColumn {
	case db.Columns.Tag.ID, db.Columns.Tag.Title, db.Columns.Tag.StatusID:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}

	return v
}

// Count returns count Tags according to conditions in search params.
//
//zenrpc:search TagSearch
//zenrpc:return int
//zenrpc:500 Internal Error
func (s TagService) Count(ctx context.Context, search *TagSearch) (int, error) {
	count, err := s.portalRepo.CountTags(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get returns а list of Tags according to conditions in search params.
//
//zenrpc:search TagSearch
//zenrpc:viewOps ViewOps
//zenrpc:return []TagSummary
//zenrpc:500 Internal Error
func (s TagService) Get(ctx context.Context, search *TagSearch, viewOps *ViewOps) ([]TagSummary, error) {
	list, err := s.portalRepo.TagsByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.portalRepo.FullTag())
	if err != nil {
		return nil, InternalError(err)
	}
	tags := make([]TagSummary, 0, len(list))
	for i := 0; i < len(list); i++ {
		if tag := NewTagSummary(&list[i]); tag != nil {
			tags = append(tags, *tag)
		}
	}
	return tags, nil
}

// GetByID returns a Tag by its ID.
//
//zenrpc:id int
//zenrpc:return Tag
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s TagService) GetByID(ctx context.Context, id int) (*Tag, error) {
	db, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewTag(db), nil
}

func (s TagService) byID(ctx context.Context, id int) (*db.Tag, error) {
	db, err := s.portalRepo.TagByID(ctx, id, s.portalRepo.FullTag())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}

// Add adds a Tag from the query.
//
//zenrpc:tag Tag
//zenrpc:return Tag
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s TagService) Add(ctx context.Context, tag Tag) (*Tag, error) {
	if ve := s.isValid(ctx, tag, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	db, err := s.portalRepo.AddTag(ctx, tag.ToDB())
	if err != nil {
		return nil, InternalError(err)
	}
	return NewTag(db), nil
}

// Update updates the Tag data identified by id from the query.
//
//zenrpc:tags Tag
//zenrpc:return Tag
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s TagService) Update(ctx context.Context, tag Tag) (bool, error) {
	if _, err := s.byID(ctx, tag.ID); err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, tag, true); ve.HasErrors() {
		return false, ve.Error()
	}

	ok, err := s.portalRepo.UpdateTag(ctx, tag.ToDB())
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete deletes the Tag by its ID.
//
//zenrpc:id int
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s TagService) Delete(ctx context.Context, id int) (bool, error) {
	if _, err := s.byID(ctx, id); err != nil {
		return false, err
	}

	ok, err := s.portalRepo.DeleteTag(ctx, id)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Validate verifies that Tag data is valid.
//
//zenrpc:tag Tag
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s TagService) Validate(ctx context.Context, tag Tag) ([]FieldError, error) {
	isUpdate := tag.ID != 0
	if isUpdate {
		_, err := s.byID(ctx, tag.ID)
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, tag, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s TagService) isValid(ctx context.Context, tag Tag, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, tag); v.HasInternalError() {
		return v
	}

	//custom validation starts here
	return v
}
