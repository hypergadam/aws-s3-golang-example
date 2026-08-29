package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

func uploadWithErrorHandling(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	filePath string,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", filePath, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file %q: %v\n", filePath, err)
		}
	}(file)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NoSuchBucket](err); ok {
			return fmt.Errorf("bucket %s does not exist: %w", bucketName, err)
		}

		if _, ok := errors.AsType[*types.NotFound](err); ok {
			return fmt.Errorf("object %s does not exist: %w", objectKey, err)
		}

		return fmt.Errorf("failed to upload file %q: %w", filePath, err)
	}

	return nil
}

func main() {
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	objectKey := uuid.New().String()

	client, err := s3client.New(ctx, region, os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if err != nil {
		panic(err)
	}

	filePath := "../../photo_2026-05-20_22-00-41.jpg"

	err = uploadWithErrorHandling(ctx, client, bucketName, objectKey, filePath)
	if err != nil {
		slog.Error("failed to upload with error handling", "error", err)
	}
}
