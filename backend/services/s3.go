package services

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client *s3.S3
	bucket string
	region string
}

const (
	// URL expiration time for uploaded files (7 days)
	URLExpirationTime = 7 * 24 * time.Hour
)

func NewS3Service(accessKey, secretKey, region, bucket string) (*S3Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Service{
		client: s3.New(sess),
		bucket: bucket,
		region: region,
	}, nil
}

func (s *S3Service) UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	// Read file content
	buffer := make([]byte, header.Size)
	if _, err := file.Read(buffer); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s/%s-%s%s", folder, time.Now().Format("20060102"), uuid.New().String(), ext)

	// Upload to S3 (private bucket)
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(buffer),
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate pre-signed URL (valid for 7 days)
	url, err := s.generatePresignedURL(filename, URLExpirationTime)
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return url, nil
}

type PDFUrls struct {
	ViewUrl     string
	DownloadUrl string
}

func (s *S3Service) UploadPDF(data []byte, filename string) (string, error) {
	key := fmt.Sprintf("brochures/%s-%s.pdf", time.Now().Format("20060102"), uuid.New().String())

	// Upload PDF to S3 (private bucket) - no ContentDisposition set on upload
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to S3: %w", err)
	}

	// Generate pre-signed URL for viewing (inline)
	url, err := s.generatePresignedURLWithDisposition(
		key,
		URLExpirationTime,
		fmt.Sprintf("inline; filename=\"%s.pdf\"", filename),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return url, nil
}

func (s *S3Service) UploadPDFWithUrls(data []byte, filename string) (*PDFUrls, error) {
	key := fmt.Sprintf("brochures/%s-%s.pdf", time.Now().Format("20060102"), uuid.New().String())

	// Upload PDF to S3 (private bucket) - no ContentDisposition set on upload
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload PDF to S3: %w", err)
	}

	// Generate pre-signed URL for viewing (inline - opens in browser)
	viewUrl, err := s.generatePresignedURLWithDisposition(
		key,
		URLExpirationTime,
		fmt.Sprintf("inline; filename=\"%s.pdf\"", filename),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate view URL: %w", err)
	}

	// Generate pre-signed URL for downloading (attachment - forces download)
	downloadUrl, err := s.generatePresignedURLWithDisposition(
		key,
		URLExpirationTime,
		fmt.Sprintf("attachment; filename=\"%s.pdf\"", filename),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &PDFUrls{
		ViewUrl:     viewUrl,
		DownloadUrl: downloadUrl,
	}, nil
}

// generatePresignedURL creates a temporary URL for accessing a private S3 object
func (s *S3Service) generatePresignedURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	// Generate pre-signed URL with expiration time
	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to create pre-signed URL: %w", err)
	}

	return url, nil
}

// generatePresignedURLWithDisposition creates a pre-signed URL with custom response headers
func (s *S3Service) generatePresignedURLWithDisposition(key string, expiration time.Duration, disposition string) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(disposition),
	})

	// Generate pre-signed URL with expiration time
	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to create pre-signed URL: %w", err)
	}

	return url, nil
}

