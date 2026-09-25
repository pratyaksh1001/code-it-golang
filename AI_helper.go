package main

import "github.com/gin-gonic/gin"

type User_data struct {
	Code     string `json:"code"`
	Question string `json:"question"`
}

func helper_bot(c *gin.Context) {
	var user_data User_data = User_data{}
	c.BindJSON(&user_data)

}
