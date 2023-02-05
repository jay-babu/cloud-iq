package aws

import (
	"encoding/json"
	"fmt"
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
	ReadMetricName  string    `json:"readMetricName"`
	WriteMetricName string    `json:"writeMetricName"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Namespace       string    `json:"namespace"`
	Period          int32     `json:"period"`
}

func DefaultAwsDdbUnused() awsDdbUnusedParams {
	return awsDdbUnusedParams{
		ReadMetricName:  "ConsumedReadCapacityUnits",
		WriteMetricName: "ConsumedWriteCapacityUnits",
		StartTime:       time.Now().AddDate(0, -3, 0),
		EndTime:         time.Now(),
		Namespace:       "AWS/DynamoDB",
		Period:          1 * 60 * 60 * 60,
	}
}

func AwsDdbUnused(ctx *gin.Context, param awsDdbUnusedParams) {
	listTablesOutput, err := ddbClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		ctx.Error(err)
		return
	}
	for _, tableName := range listTablesOutput.TableNames {
		log.SLogger.Debugf("Working on Table %s\n", tableName)
		tableDescription, err := ddbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: &tableName,
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		tableHasRead, err := isTableRead(ctx, param, tableDescription)
		if err != nil {
			ctx.Error(err)
			return
		}
		tableHasWrite, err := isTableWritten(ctx, param, tableDescription)
		if err != nil {
			ctx.Error(err)
			return
		}

		if tableHasRead && !tableHasWrite {
			// ctx.JSON(http.StatusOK, gin.H{
			// 	"message":   "Table has not been read from, but is written to. Stop writing to it to safely delete it.",
			// 	"errorType": "TableConsumesWriteNotReadCapacity",
			// })
			log.SLogger.Infof(
				"Table %s has not been read from, but is written to. Stop writing to it to safely delete it.\n",
				tableName,
			)
		} else if tableHasRead && tableHasWrite {
			// ctx.JSON(http.StatusOK, gin.H{
			// 	"message":   "Table has not been read from, but is written to. Stop writing to it to safely delete it.",
			// 	"errorType": "TableConsumesWriteNotReadCapacity",
			// })
			log.SLogger.Infof(
				"Table %s has not been read from, but is written to. Stop writing to it to safely delete it.\n",
				tableName,
			)
		} else if !tableHasRead && !tableHasWrite {
			log.SLogger.Infof(
				"Table %s has not been read or written to. Deleting Table",
				tableName,
			)
			backupName := fmt.Sprintf("%s-backup", *tableDescription.Table.TableName)
			createBackupOutput, err := ddbClient.CreateBackup(ctx, &dynamodb.CreateBackupInput{
				TableName:  &tableName,
				BackupName: &backupName,
			})
			if err != nil {
				log.SLogger.Warn(err)
				continue
			}

			b, err := json.MarshalIndent(createBackupOutput, "", "\t")
			if err != nil {
				log.SLogger.Warn(err)
				continue
			}
			log.SLogger.Debug(string(b))
			ddbClient.DescribeBackup(ctx, &dynamodb.DescribeBackupInput{BackupArn: createBackupOutput.BackupDetails.BackupArn})
			deletTableOutput, err := ddbClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
				TableName: &tableName,
			})
			if err != nil {
				log.SLogger.Warn(err)
				continue
			}
			b, err = json.MarshalIndent(deletTableOutput, "", "\t")
			if err != nil {
				log.SLogger.Warn(err)
				continue
			}
			log.SLogger.Debug(string(b))
		}
	}
}

func isTableRead(
	ctx *gin.Context,
	param awsDdbUnusedParams,
	tableDescription *dynamodb.DescribeTableOutput,
) (bool, error) {
	tableName := *tableDescription.Table.TableName

	dimensionName := "TableName"
	dimensionValue := tableDescription.Table.TableName
	tableMetricStats, err := cwClient.GetMetricStatistics(
		ctx,
		&cloudwatch.GetMetricStatisticsInput{
			MetricName: &param.ReadMetricName,
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
	if err != nil {
		return true, err
	}
	b, err := json.MarshalIndent(tableMetricStats, "", "\t")
	if err != nil {
		return true, err
	}
	log.SLogger.Debug(string(b))

	tableReadSum := 0.0
	for _, datapoint := range tableMetricStats.Datapoints {
		tableReadSum += *datapoint.Sum
	}

	if tableReadSum > 0 {
		log.SLogger.Infof(
			"%s is %.0f for Table %s. Table cannot be deleted.\n",
			param.ReadMetricName,
			tableReadSum,
			tableName,
		)
		return true, nil
	} else {
		log.SLogger.Infof("%s is 0. Table %s can be safely deleted \n", param.ReadMetricName, tableName)
		return false, nil
	}
}

func isTableWritten(
	ctx *gin.Context,
	param awsDdbUnusedParams,
	tableDescription *dynamodb.DescribeTableOutput,
) (bool, error) {
	tableName := *tableDescription.Table.TableName

	dimensionName := "TableName"
	dimensionValue := tableDescription.Table.TableName
	tableMetricStats, err := cwClient.GetMetricStatistics(
		ctx,
		&cloudwatch.GetMetricStatisticsInput{
			MetricName: &param.WriteMetricName,
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
	if err != nil {
		return true, err
	}
	b, err := json.MarshalIndent(tableMetricStats, "", "\t")
	if err != nil {
		return true, err
	}
	log.SLogger.Debug(string(b))

	tableReadSum := 0.0
	for _, datapoint := range tableMetricStats.Datapoints {
		tableReadSum += *datapoint.Sum
	}

	if tableReadSum > 0 {
		log.SLogger.Infof(
			"%s is %.0f for Table %s. Table cannot be deleted.\n",
			param.WriteMetricName,
			tableReadSum,
			tableName,
		)
		return true, nil
	} else {
		log.SLogger.Infof("%s is 0. Table %s can be safely deleted \n", param.WriteMetricName, tableName)
		return false, nil
	}
}
