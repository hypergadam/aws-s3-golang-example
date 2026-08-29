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
- `11-production-ready-service/` - A complete production-ready service incorporating `transfermanager`, file validation, and encryption.

The shared AWS S3 client setup is located in `internal/s3client/client.go`.

## Prerequisites

- Go 1.20+
- An AWS Account
- Valid AWS credentials with S3 access

## Setup & Authentication

To ensure consistent behavior across environments and avoid locally variant configurations (such as `~/.aws/credentials`), this project explicitly requires AWS credentials to be provided via environment variables.

### Using the `.env` file (or Environment Variables)
You can provide your credentials directly in the `.env` file or export them to your environment. The application will read them and inject them directly into the S3 client configuration. 

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
   Open `.env` and configure your settings:
   ```ini
   AWS_S3_REGION=us-east-1
   AWS_S3_BUCKET_NAME=my-app-bucket
   AWS_ACCESS_KEY_ID=your_access_key
   AWS_SECRET_ACCESS_KEY=your_secret_key
   ```

> [!CAUTION]
> **NEVER push your `.env` file or hardcoded credentials to GitHub!** 
> If you expose your AWS Access Keys publicly, malicious bots will scrape them in seconds and can rack up massive bills on your AWS account. This project includes a `.gitignore` file that ignores the `.env` file by default to protect you.

3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

## Usage

You can test any of the upload strategies by running the specific example. For instance:

```bash
go run ./examples/01-simple-upload
```

## Docker

The `Dockerfile` builds a single example into a minimal Alpine image that runs as a non-root user. Pick the example with the `EXAMPLE` build argument (defaults to `11-production-ready-service`):

```bash
docker build -t aws-s3-upload-backend .
docker build --build-arg EXAMPLE=06-presigned-url -t aws-s3-upload-backend:presign .
```

Credentials are passed in at runtime — never bake them into the image:

```bash
docker run --rm --env-file .env aws-s3-upload-backend
```

Examples that read or write local sample files (e.g. `01-simple-upload`, `11-production-ready-service`) need that file mounted into the container:

```bash
docker run --rm --env-file .env -v "$PWD/photo.jpg:/home/app/photo.jpg:ro" aws-s3-upload-backend
```

## CI/CD

`.github/workflows/docker-publish.yml` builds the image for `linux/amd64` and `linux/arm64` and pushes it to the GitHub Container Registry at `ghcr.io/<owner>/<repo>`. It needs no secrets — it authenticates with the built-in `GITHUB_TOKEN`.

| Event | Tags pushed |
| --- | --- |
| Push to `main` | `sha-<short>`, `sha-<full>`, `main`, `latest` |
| Push of a `v*.*.*` tag | `sha-<short>`, `sha-<full>`, `1.2.3`, `1.2` |
| Pull request | none — image is built to verify it compiles, but not pushed |

Every push is tagged with the commit SHA, so a deployment can always be pinned to the exact commit it came from:

```bash
docker pull ghcr.io/hypergadam/aws-s3-golang-example:sha-394ce85
```

The workflow also publishes an SBOM and a signed build provenance attestation for each image, verifiable with:

```bash
gh attestation verify oci://ghcr.io/hypergadam/aws-s3-golang-example:latest -R hypergadam/aws-s3-golang-example
```

> [!NOTE]
> Packages published to GHCR are private by default. After the first successful run, open the package page on GitHub and set its visibility to public if you want it pullable without authentication.

## Security & Best Practices

- **Credentials:** Refer to the caution above. We explicitly inject static credentials into the S3 client to avoid relying on locally variant configurations. In production, ensure these environment variables are injected securely.
- **Multipart Upload:** Use `transfermanager` for files larger than a few megabytes to improve stability, reliability, and speed.
- **Presigned URLs:** Rely on Presigned URLs to offload large file transfers from your backend server to S3 directly, saving bandwidth and processing power.

For more detailed security practices implemented in this repository, read [SECURITY_CONSIDERATIONS.md](SECURITY_CONSIDERATIONS.md).
