package src

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	dsn := "postgresql://postgres.ghbszyzyfsyurcgvxske:.zwRn+CVB6bjy4C@aws-0-eu-west-1.pooler.supabase.com:5432/postgres?sslmode=require"

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = DB.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println(" Connecté à PostgreSQL")
}
