package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://postgres:mysecretpassword@localhost:5432/example")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var greeting string
	err = db.QueryRow(`select 'hello world'`).Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "BeginTx failed: %v\n", err)
		os.Exit(1)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
		}
	}(tx)

	_, err = tx.Exec(`UPDATE users SET name = 'Bob' WHERE id = 1`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec failed: %v\n", err)
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Commit failed: %v\n", err)
		os.Exit(1)
	}
}
