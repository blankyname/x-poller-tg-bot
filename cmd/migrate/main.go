package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required")
	}
	dir := "migrations"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, `create table if not exists schema_migrations (version text primary key, applied_at timestamptz not null default now())`); err != nil {
		panic(err)
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		panic(err)
	}
	sort.Strings(paths)
	for _, path := range paths {
		version := filepath.Base(path)
		var exists bool
		if err := conn.QueryRow(ctx, `select exists(select 1 from schema_migrations where version=$1)`, version).Scan(&exists); err != nil {
			panic(err)
		}
		if exists {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			panic(err)
		}
		if _, err := tx.Exec(ctx, string(b)); err != nil {
			_ = tx.Rollback(ctx)
			panic(err)
		}
		if _, err := tx.Exec(ctx, `insert into schema_migrations(version) values($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			panic(err)
		}
		if err := tx.Commit(ctx); err != nil {
			panic(err)
		}
		fmt.Println("applied", version)
	}
}
