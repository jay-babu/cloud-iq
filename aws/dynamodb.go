package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"

	"github.com/jay-babu/ironMaiden/log"
)

func init() {
}

type awsDdbUnusedParams struct {
	MetricName string    `json:"metricName"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	Namespace  string    `json:"namespace"`
	Period     int32     `json:"period"`
}

func DefaultAwsDdbUnused() awsDdbUnusedParams {
	return awsDdbUnusedParams{
		MetricName: "ConsumedReadCapacityUnits",
		StartTime:  time.Now().AddDate(0, -3, 0),
		EndTime:    time.Now(),
		Namespace:  "AWS/DynamoDB",
		Period:     1 * 60 * 60 * 60,
	}
}

func AwsDdbUnused(ctx *gin.Context, param awsDdbUnusedParams) {
	listTablesOutput, err := ddbClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		ctx.Error(err)
		return
	}
	for _, tableName := range listTablesOutput.TableNames {
		log.SLogger.Debugf("Working on Table %s", tableName)
		tableDescription, err := ddbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: &tableName,
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		dimensionName := "TableName"
		dimensionValue := tableDescription.Table.TableName
		tableMetricStats, err := cwClient.GetMetricStatistics(
			ctx,
			&cloudwatch.GetMetricStatisticsInput{
				MetricName: &param.MetricName,
				StartTime:  &param.StartTime,
				EndTime:    &param.EndTime,
				Statistics: []types.Statistic{types.StatisticSum},
				Namespace:  &param.Namespace,
				Period:     &param.Period,
				Dimensions: []types.Dimension{
					{
						Name:  &dimensionName,
						Value: dimensionValue,
					},
				},
			},
		)
		log.SLogger.Debug("line 68")
		if err != nil {
			ctx.Error(err)
			return
		}
		log.SLogger.Debug("line 72")
		b, err := json.MarshalIndent(tableMetricStats, "", "\t")
		if err != nil {
			ctx.Error(err)
			return
		}
		log.SLogger.Debug(string(b))
	}
}
