package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	cfg    aws.Config
	client *cloudwatchlogs.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile("default"),
	)
	client = cloudwatchlogs.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
}
