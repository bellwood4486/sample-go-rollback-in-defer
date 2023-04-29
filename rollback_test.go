package sample_db_tx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const DATABASEURL = "postgres://postgres:mysecretpassword@localhost:5432/example"
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
			name:   "commit",
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

	db, err := sql.Open("pgx", DATABASEURL)
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
			name:   "commit",
			commit: true,
			wantErr: func(t *testing.T, err error) {
				assertErrIsTarget(t, err, pgx.ErrTxClosed)
			},
		},
		{
			name:    "no commit",
			commit:  false,
			wantErr: assertErrIsNil,
		},
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASEURL)
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
