# Security Considerations

Security should be a priority from day one when dealing with file uploads to AWS S3. Below are the key security practices referenced from the tutorial:

## 1. Use Server-Side Encryption

Ensure that data is encrypted at rest by forcing server-side encryption when uploading objects:

```go
_, err = client.PutObject(ctx, &s3.PutObjectInput{
    Bucket:               aws.String(bucketName),
    Key:                  aws.String(objectKey),
    Body:                 file,
    ServerSideEncryption: types.ServerSideEncryptionAes256,
})
```

## 2. Validate File Types

Never trust the file extension alone. Always read the first 512 bytes of a file to detect its actual MIME type using `http.DetectContentType` before allowing the upload.

```go
package main

import (
    "fmt"
    "net/http"
    "os"
)

var allowedMimeTypes = map[string]bool{
    "image/jpeg":      true,
    "image/png":       true,
    "image/gif":       true,
    "application/pdf": true,
}

func validateFileType(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Read first 512 bytes for MIME detection
    buffer := make([]byte, 512)
    _, err = file.Read(buffer)
    if err != nil {
        return err
    }

    contentType := http.DetectContentType(buffer)
    if !allowedMimeTypes[contentType] {
        return fmt.Errorf("file type %s not allowed", contentType)
    }

    return nil
}
```

## 3. Set Appropriate ACLs

Never use `public-read` ACLs unless absolutely necessary. Rely on bucket policies and IAM roles for granular access control:

```go
_, err = client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String(bucketName),
    Key:    aws.String(objectKey),
    Body:   file,
    ACL:    types.ObjectCannedACLPrivate, // Default is private
})
```

## 4. Implement File Size Limits

Prevent malicious or accidental uploads of excessively large files that could exhaust your resources or inflate your AWS bill.

```go
func checkFileSize(filePath string, maxSize int64) error {
    info, err := os.Stat(filePath)
    if err != nil {
        return err
    }
    
    if info.Size() > maxSize {
        return fmt.Errorf("file size %d exceeds limit %d", info.Size(), maxSize)
    }
    
    return nil
}
```
