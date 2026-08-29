package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

func uploadFromMemory(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	data []byte,
	contentType string,
) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object from memory: %w", err)
	}

	return nil
}

func main() {
	// Load .env from either current dir or root dir
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	_ = uuid.New().String()

	client, err := s3client.New(ctx, region, os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if err != nil {
		panic(err)
	}

	_ = "/home/hypergadam/Desktop/photo_2026-05-20_22-00-41.jpg"

	data := []byte("Hello, S3 from memory!")
	err = uploadFromMemory(ctx, client, bucketName, "memory-test.txt", data, "text/plain")
	if err != nil {
		slog.Error("failed to upload from memory", "error", err)
	} else {
		fmt.Println("Successfully uploaded from memory")
	}
}
