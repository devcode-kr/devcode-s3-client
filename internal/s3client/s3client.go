package s3client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"dosc/internal/bookmarks"
)

// S3Client is a wrapper around the AWS S3 client.
type S3Client struct {
	*s3.S3
}

// New creates a new S3 client from a bookmark.
func New(b bookmarks.Bookmark) (*S3Client, error) {
	creds := credentials.NewStaticCredentials(b.ClientKey, b.SecretKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(b.Region),
		Endpoint:         aws.String(b.Endpoint),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &S3Client{S3: s3.New(sess)}, nil
}

// ListBuckets lists all buckets.
func (c *S3Client) ListBuckets() ([]string, error) {
	result, err := c.S3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var bucketNames []string
	for _, b := range result.Buckets {
		bucketNames = append(bucketNames, aws.StringValue(b.Name))
	}
	return bucketNames, nil
}

// ListObjects lists all objects in a bucket with a given prefix.
func (c *S3Client) ListObjects(bucket, prefix string) ([]string, error) {
	result, err := c.S3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	var objectKeys []string
	for _, item := range result.Contents {
		objectKeys = append(objectKeys, aws.StringValue(item.Key))
	}
	return objectKeys, nil
}