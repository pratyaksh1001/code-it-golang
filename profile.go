package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Submission struct {
	Runtime     int       `json:"runtime"`
	SubmittedAt time.Time `json:"submitted_at"`
	Email       string    `json:"email"`
	Qid         int       `json:"qid"`
}

func get_profile(c *gin.Context) {
	var data struct {
		Token string `json:"token"`
	} = struct {
		Token string `json:"token"`
	}{}
	c.ShouldBindJSON(&data)
	token, _ := jwt.Parse(data.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})
	claims := token.Claims.(jwt.MapClaims)
	var email string
	email = claims["email"].(string)
	fmt.Println(email)
	res, _ := db.Query(context.Background(), "select runtime,submitted_at,qid from submissions where email=$1 order by submitted_at desc ;", email)
	var result []Submission = []Submission{}
	defer res.Close()
	for res.Next() {
		var row Submission = Submission{}
		row.Email = string(email)
		res.Scan(&row.Runtime, &row.SubmittedAt, &row.Qid)
		result = append(result, row)
	}
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}
