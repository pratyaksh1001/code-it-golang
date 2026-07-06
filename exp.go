package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

//var db *pgxpool.Pool

func database_con_2() {
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

// func main() {

// 	var data struct {
// 		Driver   string `json:"driver"`
// 		Solution string `json:"solution"`
// 		Qid      int    `json:"qid"`
// 	}

// 	database_con_2()
// 	src, _ := os.CreateTemp("", "*.go")
// 	defer src.Close()
// 	defer os.Remove(src.Name())
// 	db.QueryRow(context.Background(), "select code,solution from driver_go where qid=$1;", data.Qid).Scan(&data.Driver, &data.Solution)
// 	sol := data.Driver + strings.Trim(data.Solution, "package main")
// 	fmt.Println(sol)
// 	src.WriteString(sol)
// 	res := exec.Command("go", "run", src.Name())
// 	res.Stdin = strings.NewReader("8\n 1 2 3 4 5 6 7 8\n4")
// 	out, _ := res.CombinedOutput()
// 	temp := strings.Trim(strings.Trim(string(out), "\n"), " ")
// 	fmt.Println(string(temp))

// 	fmt.Println(temp == "3")

// }
