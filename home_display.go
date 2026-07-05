package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIRes struct {
	Items []struct {
		Tags  []string `json:"tags"`
		Link  string   `json:"link"`
		Title string   `json:"title"`
	} `json:"items"`
}

func get_facts() {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get("https://api.stackexchange.com/2.3/questions?order=desc&sort=votes&site=stackoverflow")
	if err != nil {
		fmt.Println("error calling API")
	}
	defer resp.Body.Close()
	body_bytes, _ := io.ReadAll(resp.Body)
	res := APIRes{}
	json.Unmarshal(body_bytes, &res)
	for _, item := range res.Items {
		fmt.Println(item)
		b, _ := json.Marshal(item)
		cache.SAdd(context.Background(), "Articles", b)
	}

}

func send_random_fact(c *gin.Context) {
	var article struct {
		Tags  []string `json:"tags"`
		Link  string   `json:"link"`
		Title string   `json:"title"`
	}

	val, err := cache.SRandMember(context.Background(), "Articles").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := json.Unmarshal([]byte(val), &article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": article,
	})
}
