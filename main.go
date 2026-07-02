package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var cache *redis.Client

func main() {
	r := gin.Default()
	opt, _ := redis.ParseURL("rediss://default:gQAAAAAAAhGlAAIgcDIyNGVmN2Y2NGQyOWM0MTRmOWUwMWI1Yzg0MzM2NzE4Mg@vital-mink-135589.upstash.io:6379")
	cache = redis.NewClient(opt)
	godotenv.Load()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
			os.Getenv("FRONTEND_URL"),
		},
		AllowMethods: []string{"GET", "PUT", "POST", "DLETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: false,
		ExposeHeaders: []string{
			"Content-Length",
		},
	}))

	database_con()
	r.POST("/register", register)
	r.POST("/login", login)
	r.POST("/auth", authenticate)
	r.POST("/run_go", run_tests)
	r.POST("/run_py", run_tests_py)
	r.POST("/question", create_question)
	r.POST("/testcase", create_testcase)
	r.POST("/question_list", question_list)

	r.GET("/tags", get_tags)
	r.GET("/problem/:qid", get_question)

	r.Run(":8000")
	defer db.Close()
}
