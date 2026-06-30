package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func create_testcase(c *gin.Context) {
	var testcase TestCase
	c.ShouldBindJSON(&testcase)
	fmt.Println(testcase)
	db.Exec(c.Request.Context(), "insert into testcases(qid,input,output) values($1,$2,$3);", testcase.Qid, testcase.Input, testcase.Output)
	c.JSON(http.StatusOK, gin.H{"created": true})
}
