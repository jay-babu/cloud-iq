package aws

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	cfg          aws.Config
	cwLogsClient *cloudwatchlogs.Client
	cwClient     *cloudwatch.Client
	// ddbClient    *dynamodb.Client
	stsClient *sts.Client
	account   *string
)

func init() {
	profile := "default"
	if name, ok := os.LookupEnv("PROFILE"); ok {
		profile = name
	}
	cfg1, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	cfg = cfg1
	stsClient = sts.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
	res, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		panic(err)
	}
	account = res.Account

	cwClient = cloudwatch.NewFromConfig(cfg)
	cwLogsClient = cloudwatchlogs.NewFromConfig(cfg)
	// ddbClient = dynamodb.NewFromConfig(cfg)
}
