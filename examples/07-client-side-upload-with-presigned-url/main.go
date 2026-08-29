package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

func uploadToPresignedURL(ctx context.Context, presignedURL, filePath string) error {
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

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedURL, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	// Set content type (so important, this one simple example)
	httpReq.Header.Set("Content-Type", "image/jpeg")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make http request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close http request body: %v\n", err)
		}
	}(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read http response body: %w", err)
		}

		return fmt.Errorf("upload failed with http status %d: %s", httpResp.StatusCode, string(body))
	}

	fmt.Println("File uploaded successfully via presigned URL")

	return nil
}

func main() {
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	_ = os.Getenv("AWS_S3_BUCKET_NAME")
	_ = uuid.New().String()

	_, err := s3client.New(ctx, region, os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if err != nil {
		panic(err)
	}

	filePath := "../../photo_2026-05-20_22-00-41.jpg"

	url := "https://example-presigned-url.com"
	err = uploadToPresignedURL(ctx, url, filePath)
	if err != nil {
		slog.Error("failed to upload", "error", err)
	}
}
