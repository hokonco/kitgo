package kitgo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hokonco/kitgo"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/gomega"
)

func Test_client_sql(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	func() {
		defer kitgo.RecoverWith(func(recv interface{}) {
			Expect(fmt.Sprintf("%v", recv)).To(Equal(`sql: unknown driver "-" (forgotten import?)`))
		})
		kitgo.SQL.New(&kitgo.SQLConfig{DataSourceName: "-", DriverName: "-"})
	}()

	kitgo.SQL.New(&kitgo.SQLConfig{DataSourceName: "file::memory:?cache=shared", DriverName: "sqlite3"})

	ctx := context.TODO()
	t.Run("fail to begin", func(t *testing.T) {
		wrap, mock := kitgo.SQL.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		tx, err := wrap.Begin()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("all expectations were already fulfilled, call to database transaction Begin was not expected"))
		Expect(tx).To(BeNil())
	})
	t.Run("stmt.ExecContext", func(t *testing.T) {
		t.Parallel()
		wrap, mock := kitgo.SQL.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		mock.ExpectBegin()
		tx, err := wrap.Begin()
		Expect(err).To(BeNil())
		Expect(tx).NotTo(BeNil())

		func() {
			stmt := wrap.Statement(ctx, tx, "")
			_, err = stmt.Exec(ctx)
			_, err = stmt.Query(ctx)
			_, err = stmt.QueryRow(ctx)
		}()

		mock.ExpectPrepare("insert into my_table_0 \\(a,b\\) values \\(1,2\\)") // 1. parent stmt (db)
		mock.ExpectPrepare("insert into my_table_0 \\(a,b\\) values \\(1,2\\)") // 2. nested stmt (tx)
		stmt := wrap.Statement(ctx, tx, "insert into my_table_0 (a,b) values (1,2)")
		// Expect(stmt.Err()).To(BeNil())
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
		wrap, mock := kitgo.SQL.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		rows := mock.NewRows("a", "b").AddRow(1.0, "1").AddRow(2.0, "2")
		mock.ExpectPrepare("select \\* from my_table_1")
		mock.ExpectQuery("select \\* from my_table_1").WillReturnRows(rows)
		resQuery, err := wrap.Statement(ctx, nil, "select * from my_table_1").Query(ctx)
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
		wrap, mock := kitgo.SQL.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		rows := mock.NewRows("a", "b").AddRow(1.0, "1").AddRow(2.0, "2")
		mock.ExpectPrepare("select \\* from my_table_2")
		mock.ExpectQuery("select \\* from my_table_2").WillReturnRows(rows)
		resQueryRow, err := wrap.Statement(ctx, nil, "select * from my_table_2").QueryRow(ctx)
		Expect(err).To(BeNil())
		Expect(resQueryRow).To(HaveLen(2))
		Expect(resQueryRow["a"]).To(Equal(1.0))
		Expect(resQueryRow["b"]).To(Equal("1"))
	})
	t.Run("sqlC.Done", func(t *testing.T) {
		t.Parallel()
		wrap, mock := kitgo.SQL.Test()
		defer func() { Expect(mock.ExpectationsWereMet()).To(BeNil()) }()

		func() {
			mock.ExpectBegin()
			tx, err := wrap.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectCommit()
			err = wrap.Done(tx, nil)
			Expect(err).To(BeNil())
		}()
		func() {
			mock.ExpectBegin()
			tx, err := wrap.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectRollback()
			err = wrap.Done(tx, fmt.Errorf("error"))
			Expect(err).NotTo(BeNil())
		}()
		func() {
			mock.ExpectBegin()
			tx, err := wrap.Begin()
			Expect(err).To(BeNil())
			Expect(tx).NotTo(BeNil())
			mock.ExpectRollback().WillReturnError(fmt.Errorf("error"))
			err = wrap.Done(tx, fmt.Errorf("error"))
			Expect(err).NotTo(BeNil())
		}()
	})
}
