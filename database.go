package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var db *pgxpool.Pool

func database_con() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(context.Background(), connStr)
	db = conn
	//db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users ( id SERIAL PRIMARY KEY, username TEXT NOT NULL, email TEXT NOT NULL UNIQUE, password TEXT NOT NULL );")
	//db.Close(context.Background())
	//db.Exec(context.Background(), "drop table question;")
	//db.Exec(context.Background(), "create table if not exists question (qid serial primary key,email text,problem text, tags text[]);")
	//db.Exec(context.Background(), "create table testcases (tid serial primary key,qid int,input text,output text);")

}
