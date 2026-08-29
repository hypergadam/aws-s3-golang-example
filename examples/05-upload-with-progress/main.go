package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

type ProgressReader struct {
	file       *os.File
	totalBytes int64
	readBytes  *atomic.Int64
	onProgress func(uploaded, total int64)
}

var _ io.Reader = (*ProgressReader)(nil)

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.file.Read(p)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", pr.file.Name(), err)
	}

	if n > 0 {
		uploaded := pr.readBytes.Add(int64(n))
		if pr.onProgress != nil {
			pr.onProgress(uploaded, pr.totalBytes)
		}
	}

	return n, nil
}

func uploadWithProgress(
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

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	totalBytes := fileInfo.Size()
	var readBytes atomic.Int64
	progressReader := &ProgressReader{
		file:       file,
		totalBytes: totalBytes,
		readBytes:  &readBytes,
		onProgress: func(uploaded, total int64) {
			percentage := float64(uploaded) / float64(total) * 100
			fmt.Printf("\rProgress: %.2f%% ( %d/%d bytes)", percentage, uploaded, total)
		},
	}

	uploader := transfermanager.New(client, func(o *transfermanager.Options) {
		o.PartSizeBytes = 10 * 1024 * 1024
		o.Concurrency = 5
	})

	// upload with process tracking
	_, err = uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   progressReader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s with process tracking: %w", filePath, err)
	}

	fmt.Println("\nCompleted successfully")
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

	err = uploadWithProgress(ctx, client, bucketName, objectKey, filePath)
	if err != nil {
		slog.Error("failed to upload with progress", "error", err)
	}
}
