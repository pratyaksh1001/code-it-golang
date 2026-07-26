package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func create_testcase(c *gin.Context) {
	var testcase struct {
		Input  string `json:"input"`
		Output string `json:"output"`
		Token  string `json:"token"`
		Qid    int    `json:"qid"`
	}
	c.ShouldBindJSON(&testcase)
	fmt.Println(testcase)
	token, _ := jwt.Parse(testcase.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})
	claims := token.Claims.(jwt.MapClaims)
	email, _ := claims["email"]
	fmt.Println()
	var solution_code_go string
	var driver_code_go string
	var imports_code_go string
	db.QueryRow(context.Background(), "select main,solution,imports from driver_go where qid=$1;", testcase.Qid).Scan(&driver_code_go, &solution_code_go, &imports_code_go)
	src, _ := os.CreateTemp("", "*.go")
	defer os.Remove(src.Name())

	solution := imports_code_go + "\n" + solution_code_go + "\n" + driver_code_go
	fmt.Println(solution)
	_, err := src.WriteString(solution)
	if err != nil {
		fmt.Println("error occured while writing in temp file")
	}
	fmt.Println(solution_code_go)
	src.Close()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*3))
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", src.Name())
	cmd.Stdin = strings.NewReader(testcase.Input)
	res, _ := cmd.CombinedOutput()
	actual := strings.Trim(strings.TrimSpace(string(res)), "\n")
	expected := strings.Trim(strings.TrimSpace(string(testcase.Output)), "\n")
	flag := true
	fmt.Printf("Actual   : %q\n", actual)
	fmt.Printf("Expected : %q\n", expected)
	fmt.Printf("Equal    : %v\n", actual == expected)
	if actual == expected {
		fmt.Println("inserted")
		db.Exec(c.Request.Context(), "insert into testcases(qid,input,output,email) values($1,$2,$3,$4);", testcase.Qid, testcase.Input, testcase.Output, email)
	} else {
		flag = false
	}
	c.JSON(http.StatusOK, gin.H{"created": flag})
}
