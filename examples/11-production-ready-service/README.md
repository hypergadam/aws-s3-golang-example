# Production-Ready Upload Service

This example ties everything together to demonstrate a robust, production-ready backend service.

It incorporates the latest `transfermanager` API (since `manager.Uploader` is deprecated) and implements all critical best practices for security and performance:
1. **Configurable Constraints:** Sets maximum file sizes and allowed MIME types.
2. **File Validation:** Safely checks the first 512 bytes of a file to detect its true `Content-Type` before uploading, preventing malicious uploads disguised by an extension.
3. **Smart Uploads:** Uses `transfermanager.Uploader` which automatically decides between standard single-part uploads and fast concurrent multi-part uploads depending on file size.
4. **Server-Side Encryption:** Explicitly uses `AES256` encryption for at-rest storage.
5. **Presigned URLs:** Contains a helper function to securely generate presigned URLs if your architecture prefers client-side uploads.

## Run

```bash
go run .
```
