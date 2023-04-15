package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gin-gonic/gin"

	"github.com/jay-babu/cloud-warden/log"
	"github.com/jay-babu/cloud-warden/oapi"
)

func DefaultAwsOldParams() oapi.AwsLogRetentionInput {
	defaultRetention := int32(180)
	return oapi.AwsLogRetentionInput{
		RetentionInDays: &defaultRetention,
	}
}

func AwsLogsOld(
	ctx *gin.Context,
	params oapi.AwsLogRetentionInput,
) (oapi.AwsLogRetentionOutput, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			aws.CredentialsProvider(credentials.NewStaticCredentialsProvider(params.AccessKeyID,
				params.SecretAccessKey, params.SessionToken)),
		),
		config.WithRegion(params.Region),
	)
	cfg.DefaultsMode = aws.DefaultsModeStandard
	cfg.RetryMode = aws.RetryModeAdaptive
	cfg.RetryMaxAttempts = 2

	if err != nil {
		return oapi.AwsLogRetentionOutput{}, err
	}
	cwLogsClient := cloudwatchlogs.NewFromConfig(cfg)

	r, err := cwLogsClient.DescribeLogGroups(ctx, nil)
	if err != nil {
		return oapi.AwsLogRetentionOutput{}, err
	}

	retention := params.RetentionInDays
	messages := make([]oapi.Message, 0, len(r.LogGroups))
	for {
		for _, l := range r.LogGroups {
			noRetentionDays := l.RetentionInDays == nil
			retentionTooLong := func() bool {
				return time.Now().
					AddDate(0, 0, -int(*l.RetentionInDays)).
					Before(time.Now().AddDate(0, 0, -(int(*retention) + 1)))
			}

			if noRetentionDays {
				log.SLogger.Infof(
					"Retention Policy for Log Group %s Does Not Exist. Setting to %d days",
					*l.Arn,
					retention,
				)
			} else if retentionTooLong() {
				log.SLogger.Infof("Retention Policy for Log Group %s is Too High: %d days. Setting to %d days", *l.Arn, *l.RetentionInDays, retention)
			}

			if noRetentionDays || retentionTooLong() {
				previousValue := int32(0)
				if !noRetentionDays {
					previousValue = *l.RetentionInDays
				}
				newValue := *retention
				arn := *l.Arn
				_, err = cwLogsClient.PutRetentionPolicy(
					ctx,
					&cloudwatchlogs.PutRetentionPolicyInput{
						LogGroupName:    l.LogGroupName,
						RetentionInDays: retention,
					},
				)
				if err != nil {
					return oapi.AwsLogRetentionOutput{}, err
				}
				messages = append(messages, oapi.Message{
					Arn:           arn,
					PreviousValue: previousValue,
					NewValue:      newValue,
					Message: fmt.Sprintf(
						"Log Group Retention Policy modified to %d days.",
						newValue,
					),
				})
			}
		}

		if r.NextToken == nil {
			break
		}

		r, err = cwLogsClient.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: r.NextToken,
		})
		if err != nil {
			return oapi.AwsLogRetentionOutput{}, err
		}
	}

	return oapi.AwsLogRetentionOutput{Messages: messages}, nil

	// if Log Group has no Log Streams and was created before a certain date
	// if len(logStreams.LogStreams) == 0 &&
	// 	time.UnixMilli(*l.CreationTime).Before(time.Now().AddDate(0, 0, 0)) {
	// 	log.SLogger.Debug(*l.LogGroupName)
	// 	deleteLogGroupInput := &cloudwatchlogs.DeleteLogGroupInput{
	// 		LogGroupName: l.LogGroupName,
	// 	}
	// 	client.PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options))
	// 	log.SLogger.Infof("Deleting Log Group %s\n", *l.LogGroupName)
	// 	deleteLogGroupOutput, err := client.DeleteLogGroup(ctx, deleteLogGroupInput)
	// 	if err != nil {
	// 		ctx.Error(err)
	// 		return
	// 	}
	// 	b, _ := json.MarshalIndent(deleteLogGroupOutput, "", "\t")
	// 	log.SLogger.Debug(string(b))
	// }
}
