package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := "postgres://postgres.evtxdcjalxiyozvkxcgu:baktiar.kuan@aws-1-ap-south-1.pooler.supabase.com:6543/postgres?sslmode=require"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), "SELECT datname FROM pg_database")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		fmt.Println(name)
	}
}
