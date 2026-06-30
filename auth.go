package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var Signature_key []byte = []byte("Pratyaksh")

func HashPass(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func CheckPass(password string, db_password string) bool {
	res := bcrypt.CompareHashAndPassword([]byte(password), []byte(db_password))
	return res == nil
}

func register(c *gin.Context) {

	log.Println("register route")
	var data User
	c.ShouldBindJSON(&data)
	fmt.Println(data)
	var (
		id       int
		username string
		email    string
		password string
	)
	err := db.QueryRow(context.Background(), "select * from users where email=$1", data.Email).Scan(&id, &username, &email, &password)
	if err != nil {
		fmt.Println(err)
	}
	if email == "" {
		ct, err := db.Exec(context.Background(), "insert into users(username,email,password) values($1,$2,$3)", data.Username, data.Email, HashPass(data.Password))
		fmt.Println(ct)
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"created": true, "exists": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"created": false, "exists": true})
	}
}

func login(c *gin.Context) {
	var data Login_data

	c.ShouldBindJSON(&data)
	var (
		email    string
		password string
		username string
	)
	err := db.QueryRow(c.Request.Context(), "select email,password,username from users where email=$1", data.Email).Scan(&email, &password, &username)
	if err != nil {
		fmt.Println(err)
	}
	pass_correct := CheckPass(password, data.Password)

	if pass_correct {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"email": email, "username": username, "exp": time.Now().Add(time.Hour * 1).Unix()})
		token_str, _ := token.SignedString(Signature_key)
		c.JSON(http.StatusOK, gin.H{"token": token_str, "username": username, "exists": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed"})
	}
}

func authenticate(c *gin.Context) {
	var data struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
		})
		return
	}

	tokenDecoded, err := jwt.Parse(data.Token, func(token *jwt.Token) (interface{}, error) {
		return Signature_key, nil
	})

	if err != nil || tokenDecoded == nil || !tokenDecoded.Valid {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
		})
		return
	}

	claims, ok := tokenDecoded.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
		})
		return
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
		})
		return
	}

	if time.Now().After(time.Unix(int64(exp), 0)) {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}
