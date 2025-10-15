package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

const (
	attachmentDirectory = "media"
)

type Client struct {
	s3     *awss3.S3
	bucket string
}

// NewClient creates S3 client
func NewClient(endpoint, accessKeyId, secretAccessKey, region, bucket string, useSSL bool) (*Client, error) {
	ep := endpoint
	if !strings.HasPrefix(strings.ToLower(ep), "http://") && !strings.HasPrefix(strings.ToLower(ep), "https://") {
		if useSSL {
			ep = "https://" + ep
		} else {
			ep = "http://" + ep
		}
	}

	cfg := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(ep),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
		DisableSSL:       aws.Bool(!useSSL),
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{s3: awss3.New(sess), bucket: bucket}, nil
}

// MakeUploadAttachment returns a presigned PUT URL to upload the object with one-minute duration.
func (c *Client) MakeUploadAttachment(ctx context.Context, channelId, objectId, fileSize int64, objectName string) (string, error) {
	key := fmt.Sprintf("%s/%d/%d/%s", attachmentDirectory, channelId, objectId, objectName)
	req, _ := c.s3.PutObjectRequest(&awss3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(fileSize),
	})
	return req.Presign(1 * time.Minute)
}

func (c *Client) RemoveAttachment(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObjectWithContext(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}
