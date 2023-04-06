package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"

	"github.com/jay-babu/auto-tune/aws"
	"github.com/jay-babu/auto-tune/log"
	"github.com/jay-babu/auto-tune/oapi"
)

type ServerImpl struct{}

var _ oapi.ServerInterface = (*ServerImpl)(nil)

func (ServerImpl) LogGroupRetention(c *gin.Context) {
	params := aws.DefaultAwsOldParams()

	_ = c.ShouldBind(&params)
	aws.AwsLogsOld(c, params)
	if l := len(c.Errors); l == 0 {
		c.Status(http.StatusNoContent)
	}
}

func main() {
	r := gin.New()
	r.Use(requestid.New())

	r.Use(ginzap.Ginzap(log.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(log.Logger, true))

	swagger, err := oapi.GetSwagger()
	if err != nil {
		// This should never error
		panic("there was an error getting the swagger")
	}

	// Clear out the servers array in the swagger spec. It is recommended to do this so that it skips validating
	// that server names match.
	swagger.Servers = nil

	r.Use(middleware.OapiRequestValidator(swagger))

	var myAPI ServerImpl

	r = oapi.RegisterHandlers(r, myAPI)
	r.Run()
}
