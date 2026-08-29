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
	"os"
)

func uploadFile(ctx context.Context, client *s3.Client, bucketName, objectKey, filePath string) error {
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

	// upload file to s3.
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %w", filePath, err)
	}

	fmt.Printf("Successfully uploaded %s to %s\n", objectKey, bucketName)
	return nil
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

	err = uploadFile(ctx, client, bucketName, objectKey, filePath)
	if err != nil {
		slog.Error("failed to upload object", "error", err)
	}
}
