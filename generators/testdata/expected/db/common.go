package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type CommonRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewCommonRepo returns new repository
func NewCommonRepo(db orm.DB) CommonRepo {
	return CommonRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.SiteUser.Name: {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.LoginCode.Name: {{Column: Columns.LoginCode.CreatedAt, Direction: SortDesc}},
			Tables.SiteUser.Name:  {{Column: Columns.SiteUser.CreatedAt, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.LoginCode.Name: {TableColumns, Columns.LoginCode.SiteUser},
			Tables.SiteUser.Name:  {TableColumns},
		},
	}
}

// WithTransaction is a function that wraps CommonRepo with pg.Tx transaction.
func (cr CommonRepo) WithTransaction(tx *pg.Tx) CommonRepo {
	cr.db = tx
	return cr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (cr CommonRepo) WithEnabledOnly() CommonRepo {
	f := make(map[string][]Filter, len(cr.filters))
	for i := range cr.filters {
		f[i] = make([]Filter, len(cr.filters[i]))
		copy(f[i], cr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	cr.filters = f

	return cr
}

/*** LoginCode ***/

// FullLoginCode returns full joins with all columns
func (cr CommonRepo) FullLoginCode() OpFunc {
	return WithColumns(cr.join[Tables.LoginCode.Name]...)
}

// DefaultLoginCodeSort returns default sort.
func (cr CommonRepo) DefaultLoginCodeSort() OpFunc {
	return WithSort(cr.sort[Tables.LoginCode.Name]...)
}

// LoginCodeByID is a function that returns LoginCode by ID(s) or nil.
func (cr CommonRepo) LoginCodeByID(ctx context.Context, id string, ops ...OpFunc) (*LoginCode, error) {
	return cr.OneLoginCode(ctx, &LoginCodeSearch{ID: &id}, ops...)
}

// OneLoginCode is a function that returns one LoginCode by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneLoginCode(ctx context.Context, search *LoginCodeSearch, ops ...OpFunc) (*LoginCode, error) {
	obj := &LoginCode{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.LoginCode.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// LoginCodesByFilters returns LoginCode list.
func (cr CommonRepo) LoginCodesByFilters(ctx context.Context, search *LoginCodeSearch, pager Pager, ops ...OpFunc) (loginCodes []LoginCode, err error) {
	err = buildQuery(ctx, cr.db, &loginCodes, search, cr.filters[Tables.LoginCode.Name], pager, ops...).Select()
	return
}

// CountLoginCodes returns count
func (cr CommonRepo) CountLoginCodes(ctx context.Context, search *LoginCodeSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &LoginCode{}, search, cr.filters[Tables.LoginCode.Name], PagerOne, ops...).Count()
}

// AddLoginCode adds LoginCode to DB.
func (cr CommonRepo) AddLoginCode(ctx context.Context, loginCode *LoginCode, ops ...OpFunc) (*LoginCode, error) {
	q := cr.db.ModelContext(ctx, loginCode)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.LoginCode.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return loginCode, err
}

// UpdateLoginCode updates LoginCode in DB.
func (cr CommonRepo) UpdateLoginCode(ctx context.Context, loginCode *LoginCode, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, loginCode).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.LoginCode.ID, Columns.LoginCode.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteLoginCode deletes LoginCode from DB.
func (cr CommonRepo) DeleteLoginCode(ctx context.Context, id string) (deleted bool, err error) {
	loginCode := &LoginCode{ID: id}

	res, err := cr.db.ModelContext(ctx, loginCode).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** SiteUser ***/

// FullSiteUser returns full joins with all columns
func (cr CommonRepo) FullSiteUser() OpFunc {
	return WithColumns(cr.join[Tables.SiteUser.Name]...)
}

// DefaultSiteUserSort returns default sort.
func (cr CommonRepo) DefaultSiteUserSort() OpFunc {
	return WithSort(cr.sort[Tables.SiteUser.Name]...)
}

// SiteUserByID is a function that returns SiteUser by ID(s) or nil.
func (cr CommonRepo) SiteUserByID(ctx context.Context, id int, ops ...OpFunc) (*SiteUser, error) {
	return cr.OneSiteUser(ctx, &SiteUserSearch{ID: &id}, ops...)
}

// OneSiteUser is a function that returns one SiteUser by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneSiteUser(ctx context.Context, search *SiteUserSearch, ops ...OpFunc) (*SiteUser, error) {
	obj := &SiteUser{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.SiteUser.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// SiteUsersByFilters returns SiteUser list.
func (cr CommonRepo) SiteUsersByFilters(ctx context.Context, search *SiteUserSearch, pager Pager, ops ...OpFunc) (siteUsers []SiteUser, err error) {
	err = buildQuery(ctx, cr.db, &siteUsers, search, cr.filters[Tables.SiteUser.Name], pager, ops...).Select()
	return
}

// CountSiteUsers returns count
func (cr CommonRepo) CountSiteUsers(ctx context.Context, search *SiteUserSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &SiteUser{}, search, cr.filters[Tables.SiteUser.Name], PagerOne, ops...).Count()
}

// AddSiteUser adds SiteUser to DB.
func (cr CommonRepo) AddSiteUser(ctx context.Context, siteUser *SiteUser, ops ...OpFunc) (*SiteUser, error) {
	q := cr.db.ModelContext(ctx, siteUser)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.SiteUser.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return siteUser, err
}

// UpdateSiteUser updates SiteUser in DB.
func (cr CommonRepo) UpdateSiteUser(ctx context.Context, siteUser *SiteUser, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, siteUser).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.SiteUser.ID, Columns.SiteUser.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteSiteUser set statusId to deleted in DB.
func (cr CommonRepo) DeleteSiteUser(ctx context.Context, id int) (deleted bool, err error) {
	siteUser := &SiteUser{ID: id, StatusID: StatusDeleted}

	return cr.UpdateSiteUser(ctx, siteUser, WithColumns(Columns.SiteUser.StatusID))
}
