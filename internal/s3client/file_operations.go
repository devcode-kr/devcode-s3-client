package s3client

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadFile uploads a file to an S3 bucket.
func (c *S3Client) UploadFile(bucket, key, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploaderWithClient(c.S3)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

// DownloadFile downloads a file from an S3 bucket.
func (c *S3Client) DownloadFile(bucket, key, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloaderWithClient(c.S3)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// DeleteObject deletes an object from an S3 bucket.
func (c *S3Client) DeleteObject(bucket, key string) error {
	_, err := c.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}