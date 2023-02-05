package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	cfg          aws.Config
	cwLogsClient *cloudwatchlogs.Client
	cwClient     *cloudwatch.Client
	ddbClient    *dynamodb.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile("default"),
	)
	if err != nil {
		panic(err)
	}

	cwClient = cloudwatch.NewFromConfig(cfg)
	cwLogsClient = cloudwatchlogs.NewFromConfig(cfg)
	ddbClient = dynamodb.NewFromConfig(cfg)
}
