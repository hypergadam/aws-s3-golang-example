package s3client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func New(_ context.Context, region, accessKey, secretKey string) (*s3.Client, error) {
	// Create S3 client bypassing LoadDefaultConfig to avoid locally variant configurations.
	client := s3.New(s3.Options{
		Region:           region,
		Credentials:      aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		RetryMaxAttempts: 3,
		RetryMode:        aws.RetryModeAdaptive,
		UsePathStyle:     true,
	})

	return client, nil
}
