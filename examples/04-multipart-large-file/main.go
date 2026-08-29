package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

func uploadLargeFile(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	filePath string,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file %s: %v\n", filePath, err)
		}
	}(file)

	uploader := transfermanager.New(client, func(o *transfermanager.Options) {
		o.PartSizeBytes = 10 * 1024 * 1024 // 10MB per part
		o.Concurrency = 5                  // Upload 5 parts concurrently
	})

	result, err := uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload large file %s: %w", filePath, err)
	}

	fmt.Printf("Successfully uploaded large file to location %s\n", *result.Location)

	return nil
}

func main() {
	// Load .env from either current dir or root dir
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	objectKey := uuid.New().String()

	client, err := s3client.New(ctx, region, os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if err != nil {
		panic(err)
	}

	filePath := "/home/hypergadam/Desktop/photo_2026-05-20_22-00-41.jpg"

	err = uploadLargeFile(ctx, client, bucketName, objectKey, filePath)
	if err != nil {
		slog.Error("failed to upload large file", "error", err)
	}
}
