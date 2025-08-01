//lint:file-ignore U1000 ignore unused code, it's generated
//nolint:structcheck,unused
package db

import (
	"context"
	"hash/crc64"
	"reflect"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// DB stores db connection
type DB struct {
	*pg.DB

	crcTable *crc64.Table
}

// New is a function that returns DB as wrapper on postgres connection.
func New(db *pg.DB) DB {
	d := DB{DB: db, crcTable: crc64.MakeTable(crc64.ECMA)}
	return d
}

// Version is a function that returns Postgres version.
func (db *DB) Version() (string, error) {
	var v string
	if _, err := db.QueryOne(pg.Scan(&v), "select version()"); err != nil {
		return "", err
	}

	return v, nil
}

// runInTransaction runs chain of functions in transaction until first error
func (db *DB) runInTransaction(ctx context.Context, fns ...func(*pg.Tx) error) error {
	return db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for _, fn := range fns {
			if err := fn(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// RunInLock runs chain of functions in transaction with lock until first error
func (db *DB) RunInLock(ctx context.Context, lockName string, fns ...func(*pg.Tx) error) error {
	lock := int64(crc64.Checksum([]byte(lockName), db.crcTable))

	return db.RunInTransaction(ctx, func(tx *pg.Tx) (err error) {
		if _, err = tx.Exec("select pg_advisory_xact_lock(?) -- ?", lock, lockName); err != nil {
			return
		}

		for _, fn := range fns {
			if err = fn(tx); err != nil {
				return
			}
		}

		return
	})
}

// buildQuery applies all functions to orm query.
func buildQuery(ctx context.Context, db orm.DB, model interface{}, search Searcher, filters []Filter, pager Pager, ops ...OpFunc) *orm.Query {
	q := db.ModelContext(ctx, model)
	for _, filter := range filters {
		filter.Apply(q)
	}

	if reflect.ValueOf(search).IsValid() && !reflect.ValueOf(search).IsNil() { // is it good?
		search.Apply(q)
	}

	q = pager.Apply(q)
	applyOps(q, ops...)

	return q
}
