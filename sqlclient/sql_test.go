package sqlclient_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/sqlclient"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_sql(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	func() {
		defer kitgo.RecoverWith(func(recv interface{}) {
			Expect(fmt.Sprint(recv)).To(Equal(`sql: unknown driver "-" (forgotten import?)`))
		})
		sqlclient.New(sqlclient.Config{DataSourceName: "-", DriverName: "-"})
	}()

	sqlclient.New(sqlclient.Config{
		DataSourceName:  "file::memory:?cache=shared",
		DriverName:      "sqlite3",
		ConnMaxIdleTime: "30s",
		ConnMaxLifetime: "30s",
	})

	ctx := context.TODO()
	t.Run("fail to begin", func(t *testing.T) {
		sqlCli, mock := sqlclient.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		tx, err := sqlCli.Begin()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("all expectations were already fulfilled, call to database transaction Begin was not expected"))
		Expect(tx).To(BeNil())
	})
	t.Run("stmt.ExecContext", func(t *testing.T) {
		t.Parallel()
		sqlCli, mock := sqlclient.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		mock.ExpectBegin()
		tx, err := sqlCli.Begin()
		Expect(err).To(BeNil())
		Expect(tx).NotTo(BeNil())

		func() {
			stmt := sqlCli.Statement(ctx, tx, "")
			_, err = stmt.Exec(ctx)
			Expect(err).To(Equal(stmt.Err()))
			_, err = stmt.Query(ctx)
			Expect(err).To(Equal(stmt.Err()))
			_, err = stmt.QueryRow(ctx)
			Expect(err).To(Equal(stmt.Err()))
		}()

		mock.ExpectPrepare("insert into my_table_0 \\(a,b\\) values \\(1,2\\)") // 1. parent stmt (db)
		mock.ExpectPrepare("insert into my_table_0 \\(a,b\\) values \\(1,2\\)") // 2. nested stmt (tx)
		stmt := sqlCli.Statement(ctx, tx, "insert into my_table_0 (a,b) values (1,2)")
		Expect(stmt.Err()).To(BeNil())
		Expect(stmt).NotTo(BeNil())

		var liv1, rav1 int64 = 0, 1
		var lie1, rae1 error = nil, nil
		mock.ExpectExec("insert into my_table_0 \\(a,b\\) values \\(1,2\\)").WillReturnResult(mock.
			NewResult(liv1, lie1, rav1, rae1),
		)
		resExec, err := stmt.Exec(ctx)
		Expect(err).To(BeNil())
		Expect(resExec.Result).NotTo(BeNil())
		liv2, lie2 := resExec.LastInsertId()
		rav2, rae2 := resExec.RowsAffected()
		Expect(liv1).To(Equal(liv2))
		if !Expect(lie1).To(BeNil()) {
			Expect(lie1).To(Equal(lie2))
		}
		Expect(rav1).To(Equal(rav2))
		if !Expect(rae1).To(BeNil()) {
			Expect(rae1).To(Equal(rae2))
		}
	})
	t.Run("stmt.QueryContext", func(t *testing.T) {
		t.Parallel()
		sqlCli, mock := sqlclient.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		rows := mock.NewRows("a", "b").AddRow(1.0, "1").AddRow(2.0, "2")
		mock.ExpectPrepare("select \\* from my_table_1")
		mock.ExpectQuery("select \\* from my_table_1").WillReturnRows(rows)
		resQuery, err := sqlCli.Statement(ctx, nil, "select * from my_table_1").Query(ctx)
		Expect(err).To(BeNil())
		Expect(resQuery).To(HaveLen(2))
		Expect(resQuery[0]).To(HaveLen(2))
		Expect(resQuery[0]["a"]).To(Equal(1.0))
		Expect(resQuery[0]["b"]).To(Equal("1"))
		Expect(resQuery[1]).To(HaveLen(2))
		Expect(resQuery[1]["a"]).To(Equal(2.0))
		Expect(resQuery[1]["b"]).To(Equal("2"))
	})
	t.Run("stmt.QueryRowContext", func(t *testing.T) {
		t.Parallel()
		sqlCli, mock := sqlclient.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		rows := mock.NewRows("a", "b").AddRow(1.0, "1").AddRow(2.0, "2")
		mock.ExpectPrepare("select \\* from my_table_2")
		mock.ExpectQuery("select \\* from my_table_2").WillReturnRows(rows)
		resQueryRow, err := sqlCli.Statement(ctx, nil, "select * from my_table_2").QueryRow(ctx)
		Expect(err).To(BeNil())
		Expect(resQueryRow).To(HaveLen(2))
		Expect(resQueryRow["a"]).To(Equal(1.0))
		Expect(resQueryRow["b"]).To(Equal("1"))
	})
	t.Run("sqlC.Done", func(t *testing.T) {
		t.Parallel()
		sqlCli, mock := sqlclient.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		func() {
			mock.ExpectBegin()
			tx, err := sqlCli.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectCommit()
			err = sqlCli.Done(tx, nil)
			Expect(err).To(BeNil())
		}()
		func() {
			mock.ExpectBegin()
			tx, err := sqlCli.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectRollback()
			err = sqlCli.Done(tx, kitgo.NewError("error"))
			Expect(err).NotTo(BeNil())
		}()
		func() {
			mock.ExpectBegin()
			tx, err := sqlCli.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectRollback().WillReturnError(kitgo.NewError("error"))
			err = sqlCli.Done(tx, kitgo.NewError("error"))
			Expect(err).NotTo(BeNil())
		}()
	})

	// Expect(mock.ExpectationsWereMet()).To(BeNil())
}
