package db

import (
	"context"
	"errors"
	"github.com/google/uuid"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type CardRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewCardRepo returns new repository
func NewCardRepo(db orm.DB) CardRepo {
	return CardRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.EncryptionKey.Name: {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.EncryptionKey.Name: {{Column: Columns.EncryptionKey.CreatedAt, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.EncryptionKey.Name: {TableColumns},
		},
	}
}

// WithTransaction is a function that wraps CardRepo with pg.Tx transaction.
func (cr CardRepo) WithTransaction(tx *pg.Tx) CardRepo {
	cr.db = tx
	return cr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (cr CardRepo) WithEnabledOnly() CardRepo {
	f := make(map[string][]Filter, len(cr.filters))
	for i := range cr.filters {
		f[i] = make([]Filter, len(cr.filters[i]))
		copy(f[i], cr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	cr.filters = f

	return cr
}

/*** EncryptionKey ***/

// FullEncryptionKey returns full joins with all columns
func (cr CardRepo) FullEncryptionKey() OpFunc {
	return WithColumns(cr.join[Tables.EncryptionKey.Name]...)
}

// DefaultEncryptionKeySort returns default sort.
func (cr CardRepo) DefaultEncryptionKeySort() OpFunc {
	return WithSort(cr.sort[Tables.EncryptionKey.Name]...)
}

// EncryptionKeyByID is a function that returns EncryptionKey by ID(s) or nil.
func (cr CardRepo) EncryptionKeyByID(ctx context.Context, id uuid.UUID, ops ...OpFunc) (*EncryptionKey, error) {
	return cr.OneEncryptionKey(ctx, &EncryptionKeySearch{ID: &id}, ops...)
}

// OneEncryptionKey is a function that returns one EncryptionKey by filters. It could return pg.ErrMultiRows.
func (cr CardRepo) OneEncryptionKey(ctx context.Context, search *EncryptionKeySearch, ops ...OpFunc) (*EncryptionKey, error) {
	obj := &EncryptionKey{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.EncryptionKey.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// EncryptionKeysByFilters returns EncryptionKey list.
func (cr CardRepo) EncryptionKeysByFilters(ctx context.Context, search *EncryptionKeySearch, pager Pager, ops ...OpFunc) (encryptionKeys []EncryptionKey, err error) {
	err = buildQuery(ctx, cr.db, &encryptionKeys, search, cr.filters[Tables.EncryptionKey.Name], pager, ops...).Select()
	return
}

// CountEncryptionKeys returns count
func (cr CardRepo) CountEncryptionKeys(ctx context.Context, search *EncryptionKeySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &EncryptionKey{}, search, cr.filters[Tables.EncryptionKey.Name], PagerOne, ops...).Count()
}

// AddEncryptionKey adds EncryptionKey to DB.
func (cr CardRepo) AddEncryptionKey(ctx context.Context, encryptionKey *EncryptionKey, ops ...OpFunc) (*EncryptionKey, error) {
	q := cr.db.ModelContext(ctx, encryptionKey)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.EncryptionKey.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return encryptionKey, err
}

// UpdateEncryptionKey updates EncryptionKey in DB.
func (cr CardRepo) UpdateEncryptionKey(ctx context.Context, encryptionKey *EncryptionKey, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, encryptionKey).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.EncryptionKey.ID, Columns.EncryptionKey.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteEncryptionKey set statusId to deleted in DB.
func (cr CardRepo) DeleteEncryptionKey(ctx context.Context, id uuid.UUID) (deleted bool, err error) {
	encryptionKey := &EncryptionKey{ID: id, StatusID: StatusDeleted}

	return cr.UpdateEncryptionKey(ctx, encryptionKey, WithColumns(Columns.EncryptionKey.StatusID))
}
