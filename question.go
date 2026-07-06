package main

import (
	"context"
	"encoding/json"
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
		Title    string   `json:"title"`
	}

	c.ShouldBindJSON(&question)
	token_decoded, _ := jwt.Parse(question.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})
	claims := token_decoded.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	tags := question.Tags
	for _, v := range tags {
		go cache.SAdd(context.Background(), "tags", v)
	}
	flag := true
	_, err := db.Exec(c.Request.Context(), "insert into question(email, problem, tags,title) values($1,$2,$3,$4);", email, question.Question, question.Tags, question.Title)
	if err != nil {
		flag = false
		fmt.Println("question creation failed")
	}
	var qid int
	db.QueryRow(c.Request.Context(), "select qid from question where title=$1;", question.Title).Scan(&qid)
	fmt.Println(qid)
	go db.Exec(context.Background(), "insert into testcases(qid,input,output,email) values($1,$2,$3,$4);", qid, question.Input, question.Output, email)
	go generate_driver_code_from_IO(qid)
	c.JSON(http.StatusOK, gin.H{
		"created": flag,
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
	fmt.Println(cache.SMembers(context.Background(), "tags").Val())
	c.JSON(http.StatusOK, gin.H{
		"tags": cache.SMembers(context.Background(), "tags").Val(),
	})
}

func get_question(c *gin.Context) {
	qid, _ := c.Params.Get("qid")

	var question struct {
		Title   string   `json:"title"`
		Problem string   `json:"problem"`
		Tags    []string `json:"tags"`
		Input   string   `json:"input"`
		Output  string   `json:"output"`
	} = struct {
		Title   string   `json:"title"`
		Problem string   `json:"problem"`
		Tags    []string `json:"tags"`
		Input   string   `json:"input"`
		Output  string   `json:"output"`
	}{}

	db.QueryRow(c.Request.Context(), "select title,problem,tags from question where qid=$1;", qid).Scan(&question.Title, &question.Problem, &question.Tags)
	db.QueryRow(c.Request.Context(), "select input,output from testcases where qid=$1 order by tid limit 1;", qid).Scan(&question.Input, &question.Output)
	fmt.Println(question)
	json_data, _ := json.Marshal(question)
	fmt.Println(string(json_data))
	c.JSON(http.StatusOK, gin.H{
		"title":       question.Title,
		"description": question.Problem,
		"input":       question.Input,
		"output":      question.Output,
		"tags":        question.Tags,
	})
}

func get_driver(c *gin.Context) {
	var data struct {
		Language string `json:"language"`
		Qid      int    `json:"qid"`
	} = struct {
		Language string `json:"language"`
		Qid      int    `json:"qid"`
	}{}
	c.ShouldBindJSON(&data)
	fmt.Println(data)
	var driver string
	var abbr string
	if data.Language == "python" {
		db.QueryRow(c.Request.Context(), "select code from driver_py where qid=$1", data.Qid).Scan(&driver)
	} else if data.Language == "go" {
		db.QueryRow(c.Request.Context(), "select code from driver_go where qid=$1", data.Qid).Scan(&driver)
	} else if data.Language == "javascript" {
		db.QueryRow(c.Request.Context(), "select code from driver_js where qid=$1", data.Qid).Scan(&driver)
	}
	db.QueryRow(c.Request.Context(), "select code from $1 where qid=$2", "driver_"+abbr, data.Qid).Scan(&driver)
	fmt.Println(driver)
	c.JSON(http.StatusOK, gin.H{
		"code": driver,
	})
}
