package kitgo

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// SQL implement a wrapper around standard "database/sql"
// and extend it to be able to support Statement and Done
var SQL sql_

type sql_ struct{}

func (sql_) New(conf *SQLConfig) *SQLWrapper {
	s, err := sql.Open(conf.DriverName, conf.DataSourceName)
	PanicWhen(err != nil || s == nil, err)
	s.SetConnMaxIdleTime(conf.ConnMaxIdleTime)
	s.SetConnMaxLifetime(conf.ConnMaxLifetime)
	s.SetMaxIdleConns(conf.MaxIdleConns)
	s.SetMaxOpenConns(conf.MaxOpenConns)
	return &SQLWrapper{DB: s}
}
func (sql_) Test() (*SQLWrapper, *SQLMock) {
	s, m, err := sqlmock.New()
	PanicWhen(err != nil || s == nil, err)
	return &SQLWrapper{DB: s}, &SQLMock{m}
}

type SQLConfig struct {
	DriverName      string
	DataSourceName  string
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}
type SQLWrapper struct {
	*sql.DB
	cache map[string]*sql.Stmt
	mutex sync.Mutex
}

// Statement will cached a prepared statement on next identical query
func (s *SQLWrapper) Statement(ctx context.Context, tx *sql.Tx, q string) SQLStatement {
	if len(s.cache) < 1 {
		s.cache = make(map[string]*sql.Stmt)
	}
	s.mutex.Lock()
	var st, ok = s.cache[q]
	if st == nil || !ok {
		var err error
		if st, err = s.PrepareContext(ctx, q); st != nil && err == nil {
			s.cache[q] = st
		}
	}
	s.mutex.Unlock()
	if tx != nil && st != nil {
		st = tx.StmtContext(ctx, st)
	}
	return SQLStatement{st}
}

// Done should validate any tx transaction with it's err, should the err
// is empty then do the Commit, if any err occured then do the Rollback
func (s *SQLWrapper) Done(tx *sql.Tx, err error) error {
	if tx != nil {
		if err != nil {
			if err_ := tx.Rollback(); err_ != nil {
				err = fmt.Errorf("%s: (%w)", err, err_)
			}
		} else {
			err = tx.Commit()
		}
	}
	return err
}

type SQLStatement struct{ stmt *sql.Stmt }

// Exec
func (s SQLStatement) Exec(ctx context.Context, args ...interface{}) (res SQLResultExec, err error) {
	if s.stmt == nil {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if sqlResult, err := s.stmt.ExecContext(ctx, args...); sqlResult != nil && err == nil {
		res.Result = sqlResult
	}
	return
}

// Query
func (s SQLStatement) Query(ctx context.Context, args ...interface{}) (rows SQLResultQuery, err error) {
	if s.stmt == nil {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var sqlRows *sql.Rows
	if sqlRows, err = s.stmt.QueryContext(ctx, args...); err == nil && sqlRows != nil {
		defer sqlRows.Close()
		var cols []string
		cols, err = sqlRows.Columns()
		for err == nil && len(cols) > 0 && sqlRows.Next() {
			if err = sqlRows.Err(); err == nil {
				vs := make([]interface{}, len(cols))
				ps := make([]interface{}, len(cols))
				for i := range vs {
					ps[i] = &vs[i]
				}
				if err = sqlRows.Scan(ps...); err == nil {
					row := make(Dict, len(cols))
					for i := range vs {
						row[cols[i]] = vs[i]
					}
					rows = append(rows, row)
				}
			}
		}
	}
	return
}

// QueryRow
func (s SQLStatement) QueryRow(ctx context.Context, args ...interface{}) (row SQLResultQueryRow, err error) {
	if s.stmt == nil {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var sqlRows *sql.Rows
	if sqlRows, err = s.stmt.QueryContext(ctx, args...); err == nil {
		defer sqlRows.Close()
		var cols []string
		cols, err = sqlRows.Columns()
		if err == nil && len(cols) > 0 && sqlRows.Next() {
			if err = sqlRows.Err(); err == nil {
				vs := make([]interface{}, len(cols))
				ps := make([]interface{}, len(cols))
				for i := range vs {
					ps[i] = &vs[i]
				}
				if err = sqlRows.Scan(ps...); err == nil {
					row = make(SQLResultQueryRow, len(cols))
					for i := range vs {
						row[cols[i]] = vs[i]
					}
				}
			}
		}
	}
	return
}

type SQLResultExec struct{ sql.Result }
type SQLResultQuery []Dict
type SQLResultQueryRow Dict

type SQLMock struct{ sqlmock.Sqlmock }

func (m SQLMock) NewRows(columns ...string) *sqlmock.Rows { return m.Sqlmock.NewRows(columns) }

func (m SQLMock) NewResult(lastInsertId int64, lastInsertIdError error, rowsAffected int64, rowsAffectedError error) sql.Result {
	return sqlResult{lastInsertId, rowsAffected, lastInsertIdError, rowsAffectedError}
}

type sqlResult struct {
	liv, rav int64
	lie, rae error
}

func (r sqlResult) LastInsertId() (int64, error) { return r.liv, r.lie }
func (r sqlResult) RowsAffected() (int64, error) { return r.rav, r.rae }
