package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func create_question(c *gin.Context) {
	var question struct {
		Token    string   `json:"token"`
		Question string   `json:"question"`
		Input    string   `json:"input"`
		Output   string   `json:"output"`
		Tags     []string `json:"tags"`
	}
	cache.SAdd(context.Background(), "tags", question.Tags)
	c.ShouldBindJSON(&question)
	token_decoded, _ := jwt.Parse(question.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})
	claims := token_decoded.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	db.Exec(c.Request.Context(), "insert into question(email, problem, tags) values($1,$2,$3)", email, question.Question, question.Tags)
	var qid int
	db.QueryRow(c.Request.Context(), "select max(qid) from question;").Scan(&qid)
	db.Exec(c.Request.Context(), "insert into testcases(qid,input,output,email) values($1,$2,$3,$4);", qid, question.Input, question.Output, email)
	c.JSON(http.StatusOK, gin.H{
		"created": true,
	})
}

func question_list(c *gin.Context) {
	var data struct {
		Query string   `json:"query"`
		Tags  []string `json:"tags"`
	}
	c.ShouldBindJSON(&data)
	fmt.Println(data.Query)
	var questions []struct {
		Qid     int      `json:"qid"`
		Tags    []string `json:"tags"`
		Title   string   `json:"title"`
		Problem string   `json:"problem"`
	}
	if data.Query == "" && len(data.Tags) == 0 {
		res, _ := db.Query(context.Background(), "select qid,tags,title,problem from question order by qid limit 10;")
		for res.Next() {
			var t struct {
				Qid     int      `json:"qid"`
				Tags    []string `json:"tags"`
				Title   string   `json:"title"`
				Problem string   `json:"problem"`
			} = struct {
				Qid     int      `json:"qid"`
				Tags    []string `json:"tags"`
				Title   string   `json:"title"`
				Problem string   `json:"problem"`
			}{}
			res.Scan(&t.Qid, &t.Tags, &t.Title, &t.Problem)
			questions = append(questions, t)
		}
		c.JSON(http.StatusOK, gin.H{
			"list": questions,
		})
		return
	}
	if data.Query != "" {
		res, _ := db.Query(
			context.Background(),
			`SELECT qid, tags, title, problem
     		FROM question
       		WHERE title ILIKE $1`,
			"%"+data.Query+"%",
		)
		for res.Next() {
			var t struct {
				Qid     int      `json:"qid"`
				Tags    []string `json:"tags"`
				Title   string   `json:"title"`
				Problem string   `json:"problem"`
			} = struct {
				Qid     int      `json:"qid"`
				Tags    []string `json:"tags"`
				Title   string   `json:"title"`
				Problem string   `json:"problem"`
			}{}
			res.Scan(&t.Qid, &t.Tags, &t.Title, &t.Problem)
			questions = append(questions, t)
		}
	}
	if len(data.Tags) != 0 {
		for _, v := range data.Tags {
			res, _ := db.Query(context.Background(), "select qid,tags,title,problem from question where $1=any(tags);", v)
			for res.Next() {
				var t struct {
					Qid     int      `json:"qid"`
					Tags    []string `json:"tags"`
					Title   string   `json:"title"`
					Problem string   `json:"problem"`
				} = struct {
					Qid     int      `json:"qid"`
					Tags    []string `json:"tags"`
					Title   string   `json:"title"`
					Problem string   `json:"problem"`
				}{}
				res.Scan(&t.Qid, &t.Tags, &t.Title, &t.Problem)
				questions = append(questions, t)
			}
		}
	}
	fmt.Println(questions)
	c.JSON(http.StatusOK, gin.H{
		"list": questions,
	})
}

func get_tags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tags": cache.SMembers(context.Background(), "tags").Val(),
	})
}
