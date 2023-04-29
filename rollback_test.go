package sample_db_tx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/jackc/pgx/v5"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

const DBConnStr = "host=localhost port=5432 user=postgres password=mysecretpassword sslmode=disable dbname=example"
const UpdataSQL = `UPDATE users SET name = 'Bob' WHERE id = 1`

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func assertErrIsNil(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected nil, but got %v", err)
	}
}

func assertErrIsTarget(t *testing.T, err error, target error) {
	if !errors.Is(err, target) {
		t.Errorf("expected %v, but got %v", target, err)
	}
}

func TestRollbackInDefer_stdlib(t *testing.T) {
	tests := []struct {
		name    string
		commit  bool
		wantErr func(*testing.T, error)
	}{
		{
			name:   "committed",
			commit: true,
			wantErr: func(t *testing.T, err error) {
				assertErrIsTarget(t, err, sql.ErrTxDone)
			},
		},
		{
			name:    "no commit",
			commit:  false,
			wantErr: assertErrIsNil,
		},
	}

	db, err := sql.Open("pgx", DBConnStr)
	failIfError(t, err)
	defer db.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := db.BeginTx(context.Background(), nil)
			failIfError(t, err)
			defer func() {
				err := tx.Rollback()
				tt.wantErr(t, err)
			}()

			_, err = tx.Exec(UpdataSQL)
			failIfError(t, err)

			if tt.commit {
				err = tx.Commit()
				failIfError(t, err)
			}
		})
	}
}

func TestRollbackInDefer_pgx(t *testing.T) {
	tests := []struct {
		name    string
		commit  bool
		wantErr func(*testing.T, error)
	}{
		{
			name:   "committed",
			commit: true,
			wantErr: func(t *testing.T, err error) {
				assertErrIsTarget(t, err, pgx.ErrTxClosed) // 自前のエラーオブジェクトと比較する必要あり
			},
		},
		{
			name:    "no commit",
			commit:  false,
			wantErr: assertErrIsNil,
		},
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DBConnStr)
	failIfError(t, err)
	defer conn.Close(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := conn.Begin(ctx)
			failIfError(t, err)
			defer func() {
				err := tx.Rollback(ctx)
				tt.wantErr(t, err)
			}()

			_, err = tx.Exec(ctx, UpdataSQL)
			failIfError(t, err)

			if tt.commit {
				err = tx.Commit(ctx)
				failIfError(t, err)
			}
		})
	}
}

func TestRollbackInDefer_gorm1_pq(t *testing.T) {
	testRollbackInDeferGorm1(t, "postgres")
}

func TestRollbackInDefer_gorm1_pgx(t *testing.T) {
	// `pgx` is not officially supported, running under compatibility mode.
	testRollbackInDeferGorm1(t, "pgx")
}

func testRollbackInDeferGorm1(t *testing.T, dialect string) {
	tests := []struct {
		name    string
		commit  bool
		wantErr func(*testing.T, error)
	}{
		{
			name:   "committed",
			commit: true,
			// 他と異なりエラーは返らない。
			// sql.ErrTxDone 以外ならエラーとして返す実装になっているため。
			// See: https://github.com/jinzhu/gorm/blob/master/main.go#L596
			wantErr: assertErrIsNil,
		},
		{
			name:    "no commit",
			commit:  false,
			wantErr: assertErrIsNil,
		},
	}

	db, err := gorm.Open(dialect, DBConnStr)
	failIfError(t, err)
	defer db.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.Begin()
			defer func() {
				err := tx.Rollback().Error
				tt.wantErr(t, err)
			}()

			err = tx.Exec(UpdataSQL).Error
			failIfError(t, err)

			if tt.commit {
				err := tx.Commit().Error
				failIfError(t, err)
			}
		})
	}
}
