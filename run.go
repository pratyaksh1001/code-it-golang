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
)

func executeGo(binary string, input string, expected string, wg *sync.WaitGroup, results chan bool) {
	defer wg.Done()

	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Runtime Error:", err)
		fmt.Println(string(out))
		results <- false
		return
	}

	actual := strings.TrimSpace(string(out))
	expected = strings.TrimSpace(expected)

	results <- (actual == expected)
}

func run_tests(c *gin.Context) {

	var data struct {
		Code     string `json:"code"`
		Qid      int    `json:"qid"`
		Token    string `json:"token"`
		Language string `json:"language"`
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
	fmt.Println("User:", claims["email"])

	rows, err := db.Query(
		c.Request.Context(),
		"SELECT input, output FROM testcases WHERE qid=$1",
		data.Qid,
	)
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
		c.JSON(http.StatusBadRequest, gin.H{
			"compile_error": string(compileOut),
		})
		return
	}

	var wg sync.WaitGroup

	results := make(chan bool)

	for _, tc := range tests {
		wg.Add(1)

		go executeGo(
			binary,
			tc.Input,
			tc.Output,
			&wg,
			results,
		)
	}

	go func() {
		wg.Wait()
		close(results)
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

	c.JSON(http.StatusOK, gin.H{
		"score":   score,
		"total":   total,
		"success": success,
	})
}

func run_tests_py(c *gin.Context) {
	var data struct {
		Code     string `json:"code"`
		Qid      int    `json:"qid"`
		Token    string `json:"token"`
		Language string `json:"language"`
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
	fmt.Println("User:", claims["email"])

	rows, err := db.Query(
		c.Request.Context(),
		"SELECT input, output FROM testcases WHERE qid=$1",
		data.Qid,
	)
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
	fmt.Println(tests)
	src, _ := os.CreateTemp("", "*.py")
	defer os.Remove(src.Name())
	src.WriteString(data.Code)
	score := 0
	total := len(tests)
	res := make(chan bool)
	var wg sync.WaitGroup
	fmt.Println(src.Name())
	now := time.Now()
	for _, v := range tests {
		wg.Add(1)
		go execute_py(src.Name(), v, res, &wg)
	}
	go func() {
		wg.Wait()
		close(res)
	}()
	for v := range res {
		if v {
			score++
		} else {
			break
		}
	}
	end := time.Now()
	time_taken := end.Sub(now)
	fmt.Println(time_taken.Milliseconds())
	passed := (score == total)
	c.JSON(http.StatusOK, gin.H{
		"status":     passed,
		"score":      score,
		"time_taken": time_taken.Milliseconds(),
	})
}

func execute_py(name string, testcase struct {
	Input  string
	Output string
}, res chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", name)
	cmd.Stdin = strings.NewReader(testcase.Input)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		res <- false
		return
	}
	if err != nil {
		log.Fatal("python code execution failed")
	}
	res <- (strings.Trim(string(out), "\n") == (testcase.Output))

}

func run_sample(c *gin.Context) {

}
