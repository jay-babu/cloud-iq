package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
		if tableDescription.Table.TableStatus != ddbTypes.TableStatusActive {
			log.SLogger.Info("Table is not active. Cannot process it right now.")
			continue
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
			log.SLogger.Infof(
				"Table %s has not been read from, but is written to. Stop writing to it to safely delete it.\n",
				tableName,
			)
		} else if tableHasRead && tableHasWrite {
			log.SLogger.Infof(
				"Table %s has not been read from, but is written to. Stop writing to it to safely delete it.\n",
				tableName,
			)
		} else if !tableHasRead && !tableHasWrite {
			log.SLogger.Infof(
				"Table %s has not been read or written to. Attempting to delete Table",
				tableName,
			)
			backUpStatus, err := isBackedUp(ctx, tableDescription)
			if err != nil {
				log.SLogger.Warn(err)
				return
			}
			if backUpStatus == InProgress {
				log.SLogger.Info("Backup in progess. Can't do anything")
				continue
			} else if backUpStatus == Complete {
				log.SLogger.Info("Backup complete. Deleting Table.")
				_, err = ddbClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
					TableName: &tableName,
				})
				if err != nil {
					log.SLogger.Warn(err)
					continue
				}
			} else if backUpStatus == NotFound {
				log.SLogger.Info("Backup not found. Creating a backup. Re-run operation once backup is complete.")
				backupName := backupName(*tableDescription.Table.TableName)
				_, err = ddbClient.CreateBackup(ctx, &dynamodb.CreateBackupInput{
					TableName:  &tableName,
					BackupName: &backupName,
				})
				if err != nil {
					log.SLogger.Warn(err)
					continue
				}
			} else {
				panic(fmt.Sprintf("Status unknown: %v", backUpStatus))
			}
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

func backupName(tableName string) string {
	return fmt.Sprintf("%s-%s", tableName, "the-janitor")
}

func isBackedUp(
	ctx *gin.Context,
	tableDescription *dynamodb.DescribeTableOutput,
) (BackedUpStatus, error) {
	backupName := backupName(*tableDescription.Table.TableName)
	backups, err := ddbClient.ListBackups(ctx, &dynamodb.ListBackupsInput{
		TableName:  tableDescription.Table.TableName,
		BackupType: ddbTypes.BackupTypeFilterUser,
	})
	if err != nil {
		return NotFound, err
	}

	for _, backup := range backups.BackupSummaries {
		if backupName == *backup.BackupName {
			switch backup.BackupStatus {
			case ddbTypes.BackupStatusCreating:
				return InProgress, nil
			case ddbTypes.BackupStatusAvailable:
				return Complete, nil
			default:
				return NotFound, nil
			}
		}
	}

	return NotFound, nil
}

type BackedUpStatus int

const (
	NotFound BackedUpStatus = iota
	InProgress
	Complete
)
