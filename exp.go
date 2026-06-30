package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var test *redis.Client

func redis_test() {
	opt, _ := redis.ParseURL("rediss://default:gQAAAAAAAhGlAAIgcDIyNGVmN2Y2NGQyOWM0MTRmOWUwMWI1Yzg0MzM2NzE4Mg@vital-mink-135589.upstash.io:6379")
	test = redis.NewClient(opt)
	arr := []string{"array", "sorting", "two pointers"}
	test.SAdd(context.Background(), "tags", arr)
}
