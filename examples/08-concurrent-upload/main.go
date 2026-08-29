package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

type UploadJob struct {
	FilePath  string
	ObjectKey string
}

type UploadResult struct {
	ObjectKey string
	Error     error
}

func uploadConcurrent(
	ctx context.Context,
	client *s3.Client,
	bucketName string,
	jobs []UploadJob,
	maxConcurrency int,
) []UploadResult {
	jobChan := make(chan UploadJob, len(jobs))
	resultChan := make(chan UploadResult, len(jobs))

	// worker pool.
	var wg sync.WaitGroup
	for range maxConcurrency {
		wg.Go(func() {
			for job := range jobChan {
				err := uploadSingleFile(ctx, client, bucketName, job.ObjectKey, job.FilePath)
				resultChan <- UploadResult{
					ObjectKey: job.ObjectKey,
					Error:     err,
				}
			}
		})
	}

	for _, job := range jobs {
		jobChan <- job
	}

	go func() {
		wg.Wait()
		close(jobChan)
	}()

	var results []UploadResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func uploadSingleFile(
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

	client, err := s3client.New(ctx, region)
	if err != nil {
		panic(err)
	}

	filePath := "../../photo_2026-05-20_22-00-41.jpg"

	jobs := []UploadJob{{FilePath: filePath, ObjectKey: objectKey}}
	results := uploadConcurrent(ctx, client, bucketName, jobs, 3)
	for _, r := range results {
		if r.Error != nil {
			slog.Error("failed to upload concurrently", "error", r.Error)
		} else {
			fmt.Println("Success:", r.ObjectKey)
		}
	}
}
