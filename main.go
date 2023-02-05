package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"github.com/jay-babu/ironMaiden/aws"
)

func main() {
	r := gin.Default()
	r.Use(requestid.New())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"time":    fmt.Sprint(time.Now().Local()),
		})
	})

	r.GET("/panic", func(ctx *gin.Context) {
		panic("Expected a panic")
	})

	r.POST("/aws/logs/old", func(ctx *gin.Context) {
		params := aws.DefaultAwsOldParams()
		ctx.BindJSON(&params)
		aws.AwsLogsOld(ctx, params)
	})
	r.Run()
}
