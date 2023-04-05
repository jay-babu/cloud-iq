package aws

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gin-gonic/gin"

	"github.com/jay-babu/cloud-cleaner/log"
	"github.com/jay-babu/cloud-cleaner/oapi"
)

func DefaultAwsOldParams() oapi.AwsLogRetentionInput {
	defaultRetention := int32(8)
	return oapi.AwsLogRetentionInput{
		RetentionInDays: &defaultRetention,
	}
}

func AwsLogsOld(ctx *gin.Context, params oapi.AwsLogRetentionInput) {
	r, err := cwLogsClient.DescribeLogGroups(ctx, nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	retention := params.RetentionInDays
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
				_, err = cwLogsClient.PutRetentionPolicy(
					ctx,
					&cloudwatchlogs.PutRetentionPolicyInput{
						LogGroupName:    l.LogGroupName,
						RetentionInDays: retention,
					},
				)
				if err != nil {
					ctx.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}

		}

		if r.NextToken == nil {
			break
		}

		r, err = cwLogsClient.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: r.NextToken,
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

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
