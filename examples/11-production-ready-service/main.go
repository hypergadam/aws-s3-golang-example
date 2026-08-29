package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	transfermanagertypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hypergadam/aws-s3-upload-tutorial/internal/s3client"
	"github.com/joho/godotenv"
)

// S3Uploader handles file uploads to S3
type S3Uploader struct {
	client     *s3.Client
	uploader   *transfermanager.Client
	bucketName string
	region     string
}

// Config holds uploader configuration
type Config struct {
	BucketName   string
	Region       string
	MaxFileSize  int64
	AllowedTypes map[string]bool
	PartSize     int64
	Concurrency  int
}

// NewS3Uploader creates a new uploader instance
func NewS3Uploader(client *s3.Client, cfg Config) *S3Uploader {
	uploader := transfermanager.New(client, func(o *transfermanager.Options) {
		if cfg.PartSize > 0 {
			o.PartSizeBytes = cfg.PartSize
		}
		if cfg.Concurrency > 0 {
			o.Concurrency = cfg.Concurrency
		}
	})

	return &S3Uploader{
		client:     client,
		uploader:   uploader,
		bucketName: cfg.BucketName,
		region:     cfg.Region,
	}
}

// UploadFile uploads a file with validation and metadata using transfermanager
func (u *S3Uploader) UploadFile(ctx context.Context, filePath, objectKey string, metadata map[string]string, cfg Config) (string, error) {
	// Validate file exists and check size
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	if cfg.MaxFileSize > 0 && info.Size() > cfg.MaxFileSize {
		return "", fmt.Errorf("file size %d exceeds limit %d", info.Size(), cfg.MaxFileSize)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}(file)

	// Detect Content Type securely via first 512 bytes
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for MIME detection: %w", err)
	}
	if n > 0 {
		detectedType := http.DetectContentType(buffer[:n])
		if len(cfg.AllowedTypes) > 0 && !cfg.AllowedTypes[detectedType] {
			return "", fmt.Errorf("file type %s not allowed", detectedType)
		}
	}

	// Reset file pointer back to start after reading bytes
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Determine general content type by extension if not binary
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Perform the upload using transfermanager which automatically handles
	// single PutObject or CreateMultipartUpload based on file size.
	result, err := u.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:               aws.String(u.bucketName),
		Key:                  aws.String(objectKey),
		Body:                 file,
		ContentType:          aws.String(contentType),
		Metadata:             metadata,
		ServerSideEncryption: transfermanagertypes.ServerSideEncryptionAes256,
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	if result.Location != nil {
		return *result.Location, nil
	}

	// Fallback if result.Location is nil
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucketName, u.region, objectKey), nil
}

// GeneratePresignedURL creates a presigned upload URL
func (u *S3Uploader) GeneratePresignedURL(ctx context.Context, objectKey, contentType string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(u.client)
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign: %w", err)
	}
	return request.URL, nil
}

func main() {
	_ = godotenv.Load(".env", "../../.env")

	ctx := context.Background()
	region := os.Getenv("AWS_S3_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")

	client, err := s3client.New(ctx, region)
	if err != nil {
		panic(err)
	}

	// Setup Production-Ready Configuration
	cfg := Config{
		BucketName:  bucketName,
		Region:      region,
		MaxFileSize: 50 * 1024 * 1024, // 50 MB limit
		AllowedTypes: map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"application/pdf": true,
		},
		PartSize:    10 * 1024 * 1024, // 10 MB per part for multipart
		Concurrency: 5,                // 5 concurrent part uploads
	}

	uploader := NewS3Uploader(client, cfg)

	filePath := "../../photo_2026-05-20_22-00-41.jpg"

	// Create a dummy file if it doesn't exist just so the example can run
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dummyFile, _ := os.Create(filePath)
		dummyFile.Write([]byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x00\x00\x01\x00\x01\x00\x00")) // Valid JPEG header
		dummyFile.Close()
		defer os.Remove(filePath)
	}

	objectKey := uuid.New().String()
	metadata := map[string]string{
		"uploaded-by": "production-service",
		"env":         "production",
	}

	// 1. Test Production File Upload
	fmt.Println("--- Testing Secure UploadFile ---")
	location, err := uploader.UploadFile(ctx, filePath, objectKey, metadata, cfg)
	if err != nil {
		slog.Error("failed to upload file", "error", err)
	} else {
		fmt.Printf("Successfully uploaded to: %s\n", location)
	}

	// 2. Test Presigned URL Generation
	fmt.Println("\n--- Testing GeneratePresignedURL ---")
	url, err := uploader.GeneratePresignedURL(ctx, "user-uploads/avatar.jpg", "image/jpeg", 15*time.Minute)
	if err != nil {
		slog.Error("failed to generate presigned url", "error", err)
	} else {
		fmt.Printf("Presigned URL (valid for 15m): %s\n", url)
	}
}
