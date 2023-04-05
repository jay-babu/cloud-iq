// Package oapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package oapi

// AwsLogRetentionInput defines model for AwsLogRetentionInput.
type AwsLogRetentionInput struct {
	RetentionInDays *int32 `json:"retentionInDays,omitempty"`
}

// LogGroupRetentionJSONRequestBody defines body for LogGroupRetention for application/json ContentType.
type LogGroupRetentionJSONRequestBody = AwsLogRetentionInput