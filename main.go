package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	"github.com/jay-babu/ironMaiden/aws"
	"github.com/jay-babu/ironMaiden/log"
)

func main() {
	r := gin.New()
	r.Use(requestid.New())

	r.Use(ginzap.Ginzap(log.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(log.Logger, true))

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
		_ = ctx.ShouldBindJSON(&params)
		aws.AwsLogsOld(ctx, params)
	})
	r.POST("/aws/ddb/unused", func(ctx *gin.Context) {
		params := aws.DefaultAwsDdbUnused()
		_ = ctx.ShouldBindJSON(&params)
		aws.AwsDdbUnused(ctx, params)
	})
	r.Run()
}
