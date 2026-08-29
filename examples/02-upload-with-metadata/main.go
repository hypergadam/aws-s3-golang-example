package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

func uploadFileWithMetadata(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	objectKey string,
	filePath string,
	metadata map[string]string,
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

	contentType := cmp.Or(mime.TypeByExtension(filepath.Ext(filePath)), "application/octet-stream")

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileInfo.Size()),
		Metadata:      metadata,
		ACL:           types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file with metadata %s: %w", filePath, err)
	}

	fmt.Printf("uploaded file with metadata %s\n", filePath)

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

	metadata := map[string]string{"uploaded-by": "tutorial-app", "version": "1.0.0"}
	err = uploadFileWithMetadata(ctx, client, bucketName, objectKey, filePath, metadata)
	if err != nil {
		slog.Error("failed to upload file with metadata", "error", err)
	}
}
