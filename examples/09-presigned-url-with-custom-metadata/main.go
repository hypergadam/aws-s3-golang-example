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

func generatePresignedUploadWithMetadata(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	contentType string,
	metadata map[string]string,
) (string, error) {
	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL with metadata: %w", err)
	}

	return request.URL, nil
}

func main() {
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	objectKey := uuid.New().String()

	client, err := s3client.New(ctx, region)
	if err != nil {
		panic(err)
	}

	filePath := "../../photo_2026-05-20_22-00-41.jpg"

	metadata := map[string]string{"user-id": "12345"}
	url, err := generatePresignedUploadWithMetadata(ctx, client, bucketName, objectKey, "image/jpeg", metadata)
	if err != nil {
		slog.Error("failed to generate presigned url", "error", err)
	} else {
		fmt.Println("Presigned URL:", url)
	}
}
