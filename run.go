package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

func executeGo(binary string, input string, expected string, wg *sync.WaitGroup, results chan bool, err_chan chan string) {
	defer wg.Done()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*3))
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Runtime Error:", err)
		fmt.Println(string(out))
		err_chan <- err.Error()
		results <- false
		fmt.Println("error occured in go testcase exec")
		return
	}

	actual := strings.TrimSpace(string(out))
	expected = strings.TrimSpace(expected)

	results <- (actual == expected)
	err_chan <- actual
}

func run_tests(c *gin.Context) {

	var data struct {
		Code       string `json:"code"`
		Qid        int    `json:"qid"`
		Token      string `json:"token"`
		Language   string `json:"language"`
		Submission bool   `json:"submission"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := jwt.Parse(data.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"]
	fmt.Println("User:", claims["email"])
	var rows pgx.Rows
	if data.Submission {
		rows, _ = db.Query(
			c.Request.Context(),
			"SELECT input, output FROM testcases WHERE qid=$1;",
			data.Qid,
		)
	} else {
		rows, _ = db.Query(
			c.Request.Context(),
			"SELECT input, output FROM testcases WHERE qid=$1 limit 1;",
			data.Qid,
		)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	type TestCase struct {
		Input  string
		Output string
	}

	var tests []TestCase

	for rows.Next() {
		var tc TestCase

		if err := rows.Scan(&tc.Input, &tc.Output); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		tests = append(tests, tc)
	}

	src, err := os.CreateTemp("", "*.go")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(src.Name())

	_, err = src.WriteString(data.Code)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	src.Close()

	bin, err := os.CreateTemp("", "solution-*")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	binary := bin.Name()
	bin.Close()
	os.Remove(binary)

	defer os.Remove(binary)

	build := exec.Command(
		"go",
		"build",
		"-o",
		binary,
		src.Name(),
	)
	compileOut, err := build.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"score":      0,
			"total":      len(tests),
			"success":    false,
			"error":      string(compileOut),
			"time_taken": 0,
		})
		return
	}

	var wg sync.WaitGroup

	results := make(chan bool, len(tests))
	err_chan := make(chan string, len(tests))
	now := time.Now()
	for _, tc := range tests {
		wg.Add(1)

		go executeGo(
			binary,
			tc.Input,
			tc.Output,
			&wg,
			results,
			err_chan,
		)
	}

	go func() {
		wg.Wait()
		close(results)
		close(err_chan)
	}()

	score := 0
	success := true
	total := len(tests)

	for passed := range results {
		if passed {
			score++
		} else {
			success = false
		}
	}
	end := time.Now()
	time_taken := end.Sub(now).Milliseconds()
	fmt.Println(time_taken)
	if score == total && data.Submission {
		db.Exec(context.Background(), "insert into submissions(qid,email,runtime,submitted_at,language) values($1,$2,$3,$4,$5);", data.Qid, email, time_taken, time.Now().UTC(), data.Language)
	}
	var console_msg string
	if !data.Submission {
		for msg := range err_chan {
			fmt.Println(msg)
			if msg != "" {
				console_msg = msg
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"score":      score,
		"total":      total,
		"success":    success,
		"error":      console_msg,
		"time_taken": time_taken,
	})
}

func run_tests_py(c *gin.Context) {
	var data struct {
		Code       string `json:"code"`
		Qid        int    `json:"qid"`
		Token      string `json:"token"`
		Language   string `json:"language"`
		Submission bool   `json:"submission"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := jwt.Parse(data.Token, func(token *jwt.Token) (any, error) {
		return Signature_key, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"]
	fmt.Println("User:", claims["email"])

	var rows pgx.Rows

	if data.Submission {
		rows, _ = db.Query(
			c.Request.Context(),
			"SELECT input, output FROM testcases WHERE qid=$1;",
			data.Qid,
		)
	} else {
		rows, _ = db.Query(
			c.Request.Context(),
			"SELECT input, output FROM testcases WHERE qid=$1 limit 1;",
			data.Qid,
		)
	}

	defer rows.Close()

	type TestCase struct {
		Input  string
		Output string
	}

	var tests []TestCase

	for rows.Next() {
		var tc TestCase

		if err := rows.Scan(&tc.Input, &tc.Output); err != nil {
			fmt.Println("error parsing IO from DB")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		tests = append(tests, tc)
	}
	fmt.Println(tests)
	src, _ := os.CreateTemp("", "*.py")
	defer os.Remove(src.Name())
	src.WriteString(data.Code)
	score := 0
	total := len(tests)
	res := make(chan bool, len(tests))
	err_chan := make(chan string, len(tests))
	var wg sync.WaitGroup
	fmt.Println(src.Name())
	now := time.Now()
	for _, v := range tests {
		wg.Add(1)
		go execute_py(src.Name(), v, res, &wg, err_chan)
	}
	go func() {
		wg.Wait()
		close(res)
		close(err_chan)
	}()
	for v := range res {
		if v {
			score++
		} else {
			break
		}
	}
	var code_failure string = ""
	if !data.Submission {
		for v := range err_chan {
			if v != "" {
				code_failure = v
			}
		}
	}

	end := time.Now()
	time_taken := end.Sub(now)
	fmt.Println(time_taken.Milliseconds())
	passed := (score == total)
	if score == total && data.Submission {
		db.Exec(context.Background(), "insert into submissions(qid,email,runtime,submitted_at,language) values($1,$2,$3,$4,$5);", data.Qid, email, time_taken, time.Now().UTC(), data.Language)
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    passed,
		"score":      score,
		"time_taken": time_taken.Milliseconds(),
		"error":      code_failure,
		"total":      total,
	})
}

func execute_py(name string, testcase struct {
	Input  string
	Output string
}, res chan bool, wg *sync.WaitGroup, err_chan chan string) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	var code_failure string = ""
	cmd := exec.CommandContext(ctx, "python3", name)
	cmd.Stdin = strings.NewReader(testcase.Input)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		code_failure = "Time limit Exceeded"
		res <- false
		err_chan <- code_failure
		return
	}
	if err != nil {
		code_failure = err.Error()
		log.Println("python code execution failed")
	}

	res <- (strings.TrimSpace(strings.Trim(string(out), "\n")) == (testcase.Output))
	if code_failure == "" {
		err_chan <- code_failure
	} else {
		err_chan <- string(out)
	}

}
