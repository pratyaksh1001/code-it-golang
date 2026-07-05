package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var test *redis.Client

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

func redis_test() {
	opt, _ := redis.ParseURL("rediss://default:gQAAAAAAAhGlAAIgcDIyNGVmN2Y2NGQyOWM0MTRmOWUwMWI1Yzg0MzM2NzE4Mg@vital-mink-135589.upstash.io:6379")
	test = redis.NewClient(opt)
	arr := []string{"array", "sorting", "two pointers"}
	test.SAdd(context.Background(), "tags", arr)
}

// func main() {
// 	r := gin.Default()

// 	r.GET("/", ws_test)
// 	r.Run(":8080")
// }
