package sample_db_tx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// ロールバックを呼んだときに ErrTxDone が返ることのテスト
func TestRollbackErrTxDone(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://postgres:mysecretpassword@localhost:5432/example")
	failIfError(t, err)
	defer db.Close()

	tx, err := db.BeginTx(context.Background(), nil)
	failIfError(t, err)
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if !errors.Is(err, sql.ErrTxDone) {
			t.Errorf("expected %v, but got %v", sql.ErrTxDone, err)
		}
	}(tx)

	_, err = tx.Exec(`UPDATE users SET name = 'Bob' WHERE id = 1`)
	failIfError(t, err)

	err = tx.Commit()
	failIfError(t, err)
}

func TestRollbackErrTxDonePGX(t *testing.T) {
	t.FailNow()
}
