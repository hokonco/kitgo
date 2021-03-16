package sqlclient

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hokonco/kitgo"
)

type Config struct {
	DriverName      string `yaml:"driver_name" json:"driver_name"`
	DataSourceName  string `yaml:"data_source_name" json:"data_source_name"`
	ConnMaxIdleTime string `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
}

func New(cfg Config) *Client {
	s, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	kitgo.PanicWhen(err != nil || s == nil, err)
	var d time.Duration
	d = kitgo.ParseDuration(cfg.ConnMaxIdleTime, time.Minute)
	s.SetConnMaxIdleTime(d)
	d = kitgo.ParseDuration(cfg.ConnMaxLifetime, time.Minute)
	s.SetConnMaxLifetime(d)
	s.SetMaxIdleConns(cfg.MaxIdleConns)
	s.SetMaxOpenConns(cfg.MaxOpenConns)
	return &Client{DB: s}
}

func Test() (*Client, Mock) {
	s, m, err := sqlmock.New()
	kitgo.PanicWhen(err != nil, err)
	return &Client{DB: s}, Mock{m}
}

type Client struct {
	mu sync.Mutex

	*sql.DB
	stmtMap map[string]*sql.Stmt
}

// Statement will cached a prepared statement on next identical query
func (s *Client) Statement(ctx context.Context, tx *sql.Tx, q string) (stmt Statement) {
	if q == "" {
		stmt.err = kitgo.NewError("invalid stmt")
		return
	}
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if len(s.stmtMap) < 1 {
			s.stmtMap = make(map[string]*sql.Stmt)
		}
		if stmt.stmt = s.stmtMap[q]; stmt.stmt == nil {
			if stmt.stmt, stmt.err = s.PrepareContext(ctx, q); stmt.err == nil && stmt.stmt != nil {
				s.stmtMap[q] = stmt.stmt
			}
		}
	}()
	if tx != nil && stmt.stmt != nil && stmt.err == nil {
		stmt.stmt = tx.StmtContext(ctx, stmt.stmt)
	}
	return
}
func (s *Client) Done(tx *sql.Tx, err error) error {
	if tx != nil {
		if err != nil {
			if err_ := tx.Rollback(); err_ != nil {
				err = kitgo.NewError(err.Error()+": (%w)", err_)
			}
		} else {
			err = tx.Commit()
		}
	}
	return err
}

type Statement struct {
	stmt *sql.Stmt
	err  error
}

func (s Statement) Err() error {
	return s.err
}
func (s Statement) Exec(ctx context.Context, args ...interface{}) (res ResultExec, err error) {
	if err = s.err; err != nil {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if sqlResult, err := s.stmt.ExecContext(ctx, args...); sqlResult != nil && err == nil {
		res.Result = sqlResult
	}
	return
}
func (s Statement) Query(ctx context.Context, args ...interface{}) (rows ResultQuery, err error) {
	if err = s.err; err != nil {
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
					row := make(kitgo.Dict, len(cols))
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
func (s Statement) QueryRow(ctx context.Context, args ...interface{}) (row ResultQueryRow, err error) {
	if err = s.err; err != nil {
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
					row = make(ResultQueryRow, len(cols))
					for i := range vs {
						row[cols[i]] = vs[i]
					}
				}
			}
		}
	}
	return
}

type ResultExec struct{ sql.Result }
type ResultQuery []kitgo.Dict
type ResultQueryRow kitgo.Dict

type Mock struct{ sqlmock.Sqlmock }

func (m Mock) NewRows(columns ...string) *sqlmock.Rows { return m.Sqlmock.NewRows(columns) }

func (m Mock) NewResult(lastInsertId int64, lastInsertIdError error, rowsAffected int64, rowsAffectedError error) sql.Result {
	return result{lastInsertId, rowsAffected, lastInsertIdError, rowsAffectedError}
}

type result struct {
	liv, rav int64
	lie, rae error
}

func (r result) LastInsertId() (int64, error) { return r.liv, r.lie }
func (r result) RowsAffected() (int64, error) { return r.rav, r.rae }

const (
	SQLLastInsertIdValue = "lastinsertid_value"
	SQLLastInsertIdError = "lastinsertid_error"
	SQLRowsAffectedValue = "rowsaffected_value"
	SQLRowsAffectedError = "rowsaffected_error"
)
