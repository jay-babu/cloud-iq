package main

import (
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	r := gin.Default()
	logger, _ := zap.NewDevelopment()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"time":    fmt.Sprint(time.Now().Local()),
		})
	})
	r.Run()
}
