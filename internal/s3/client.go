package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

const (
	attachmentDirectory = "media"
	multipartPartSize   = 8 * 1024 * 1024
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

func (c *Client) MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	req, _ := c.s3.GetObjectRequest(&awss3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return req.Presign(ttl)
}

func (c *Client) RemoveAttachment(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObjectWithContext(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// UploadObject uploads an object from a stream to S3 without requiring local disk.
func (c *Client) UploadObject(ctx context.Context, key string, body io.Reader, contentType string) (err error) {
	createIn := &awss3.CreateMultipartUploadInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		createIn.ContentType = aws.String(contentType)
	}

	created, err := c.s3.CreateMultipartUploadWithContext(ctx, createIn)
	if err != nil {
		return err
	}

	uploadID := created.UploadId
	defer func() {
		if err == nil || uploadID == nil {
			return
		}
		_, _ = c.s3.AbortMultipartUploadWithContext(context.Background(), &awss3.AbortMultipartUploadInput{
			Bucket:   aws.String(c.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
	}()

	parts := make([]*awss3.CompletedPart, 0, 4)
	for partNumber := int64(1); ; partNumber++ {
		payload, readErr := readNextUploadChunk(body, multipartPartSize)
		if readErr != nil {
			return readErr
		}
		if len(payload) == 0 {
			break
		}

		partOut, err := c.s3.UploadPartWithContext(ctx, &awss3.UploadPartInput{
			Bucket:        aws.String(c.bucket),
			Key:           aws.String(key),
			UploadId:      uploadID,
			PartNumber:    aws.Int64(partNumber),
			Body:          bytes.NewReader(payload),
			ContentLength: aws.Int64(int64(len(payload))),
		})
		if err != nil {
			return err
		}

		parts = append(parts, &awss3.CompletedPart{
			ETag:       partOut.ETag,
			PartNumber: aws.Int64(partNumber),
		})
	}

	if len(parts) == 0 {
		return fmt.Errorf("empty upload body")
	}

	_, err = c.s3.CompleteMultipartUploadWithContext(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &awss3.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	return err
}

func readNextUploadChunk(reader io.Reader, maxSize int64) ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(int(maxSize))

	_, err := io.CopyN(&buf, reader, maxSize)
	switch err {
	case nil:
		return buf.Bytes(), nil
	case io.EOF:
		return nil, nil
	case io.ErrUnexpectedEOF:
		return buf.Bytes(), nil
	default:
		return nil, err
	}
}

// StatObject performs a HEAD request to retrieve object size and content type
func (c *Client) StatObject(ctx context.Context, key string) (int64, *string, error) {
	out, err := c.s3.HeadObjectWithContext(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, nil, err
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return size, out.ContentType, nil
}

// ObjectInfo contains minimal info returned from listing
type ObjectInfo struct {
	Key  string
	Size int64
}

// ListObjectsPrefix lists object keys under a given prefix
func (c *Client) ListObjectsPrefix(ctx context.Context, prefix string, max int64) ([]ObjectInfo, error) {
	var result []ObjectInfo
	var token *string
	fetched := int64(0)
	for {
		out, err := c.s3.ListObjectsV2WithContext(ctx, &awss3.ListObjectsV2Input{
			Bucket:            aws.String(c.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
			MaxKeys:           aws.Int64(1000),
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			size := int64(0)
			if obj.Size != nil {
				size = *obj.Size
			}
			result = append(result, ObjectInfo{Key: *obj.Key, Size: size})
			fetched++
			if max > 0 && fetched >= max {
				return result, nil
			}
		}
		if out.IsTruncated == nil || !*out.IsTruncated {
			break
		}
		token = out.NextContinuationToken
	}
	return result, nil
}
