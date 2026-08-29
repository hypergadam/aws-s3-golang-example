package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
	"log/slog"
	"time"
)

func generatePresignedUploadURL(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	duration time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(client)

	// Generate presigned PUT request.
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return request.URL, nil
}

func main() {
	// Load .env from either current dir or root dir
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	objectKey := uuid.New().String()

	client, err := s3client.New(ctx, region)
	if err != nil {
		panic(err)
	}

	filePath := "/home/hypergadam/Desktop/photo_2026-05-20_22-00-41.jpg"

	url, err := generatePresignedUploadURL(ctx, client, bucketName, objectKey, 15*time.Minute)
	if err != nil {
		slog.Error("failed to generate presigned url", "error", err)
	} else {
		fmt.Printf("Presigned URL (valid for 15m): %s\n", url)
	}
}
