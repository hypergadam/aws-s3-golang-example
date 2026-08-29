# AWS S3 Upload Backend (Go SDK v2)

A comprehensive backend example demonstrating various ways to upload files to AWS S3 using the new **AWS SDK for Go v2**. 
Based on the tutorial: [How to Upload Files to AWS S3 in Go Complete Guide with SDK v2](https://buanacoding.com/2025/10/how-to-upload-files-to-aws-s3-in-go-with-sdk-v2) by BuanaCoding.

## Project Structure

This repository is divided into specific runnable examples located in the `examples/` directory. Each folder contains its own `README.md` and `main.go` file demonstrating a specific capability:

- `01-simple-upload/` - Upload files directly from disk.
- `02-upload-with-metadata/` - Attach custom metadata (e.g., tags, specific content types) to your files.
- `03-upload-from-memory/` - Upload data generated in memory without saving to disk first.
- `04-multipart-large-file/` - Upload massive files (videos, backups) concurrently using `transfermanager`.
- `05-upload-with-progress/` - Track upload percentage and bytes transferred in real-time.
- `06-presigned-url/` - Generate temporary URLs allowing clients to upload files directly to S3.
- `07-client-side-upload-with-presigned-url/` - Example of how a client uploads to a presigned URL.
- `08-concurrent-upload/` - Upload multiple files concurrently using a worker pool pattern.
- `09-presigned-url-with-custom-metadata/` - Enforce specific headers/metadata on presigned URLs.
- `10-upload-with-error-handling/` - Check for specific S3 error types (`NoSuchBucket`, etc).

The shared AWS S3 client setup is located in `internal/s3client/client.go`.

## Prerequisites

- Go 1.20+
- An AWS Account
- Valid AWS credentials with S3 access (e.g., via `~/.aws/credentials`)

## Setup

1. **Clone the repository:**
   ```bash
   git clone <your-repo-url>
   cd aws-s3-upload-backend
   ```

2. **Configure Environment Variables:**
   Copy the example environment file and fill in your details:
   ```bash
   cp .env.example .env
   ```
   *Required variables:*
   - `AWS_S3_REGION` (e.g., `us-east-1`)
   - `AWS_S3_BUCKET_NAME` (e.g., `my-app-bucket`)

3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

## Usage

You can test any of the upload strategies by running the specific example. For instance:

```bash
go run ./examples/01-simple-upload
```

## Security & Best Practices

- **Credentials:** Never hardcode your AWS credentials in the application. Rely on environment variables, `~/.aws/credentials`, or IAM roles.
- **Multipart Upload:** Use `manager.Uploader` (transfermanager) for files larger than a few megabytes to improve stability, reliability, and speed.
- **Presigned URLs:** Rely on Presigned URLs to offload large file transfers from your backend server to S3 directly, saving bandwidth and processing power.
