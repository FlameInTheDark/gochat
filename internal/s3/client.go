package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awss3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	attachmentDirectory = "media"
	multipartPartSize   = 8 * 1024 * 1024
)

type Client struct {
	s3      *awss3.Client
	presign *awss3.PresignClient
	bucket  string
}

// NewClient creates S3 client
func NewClient(endpoint, accessKeyId, secretAccessKey, region, bucket string, useSSL bool) (*Client, error) {
	cfg := aws.Config{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	}

	s3Client := awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.UsePathStyle = true
		if ep := normalizeEndpoint(endpoint, useSSL); ep != "" {
			o.BaseEndpoint = aws.String(ep)
		}
	})

	return &Client{
		s3:      s3Client,
		presign: awss3.NewPresignClient(s3Client),
		bucket:  bucket,
	}, nil
}

// MakeUploadAttachment returns a presigned PUT URL to upload the object with one-minute duration.
func (c *Client) MakeUploadAttachment(ctx context.Context, channelId, objectId, fileSize int64, objectName string) (string, error) {
	key := fmt.Sprintf("%s/%d/%d/%s", attachmentDirectory, channelId, objectId, objectName)
	req, err := c.presign.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(fileSize),
	}, awss3.WithPresignExpires(time.Minute))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (c *Client) MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	req, err := c.presign.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, awss3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (c *Client) RemoveAttachment(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &awss3.DeleteObjectInput{
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

	created, err := c.s3.CreateMultipartUpload(ctx, createIn)
	if err != nil {
		return err
	}

	uploadID := created.UploadId
	defer func() {
		if err == nil || uploadID == nil {
			return
		}
		_, _ = c.s3.AbortMultipartUpload(context.Background(), &awss3.AbortMultipartUploadInput{
			Bucket:   aws.String(c.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
	}()

	parts := make([]awss3types.CompletedPart, 0, 4)
	for partNumber := int32(1); ; partNumber++ {
		payload, readErr := readNextUploadChunk(body, multipartPartSize)
		if readErr != nil {
			return readErr
		}
		if len(payload) == 0 {
			break
		}

		partOut, err := c.s3.UploadPart(ctx, &awss3.UploadPartInput{
			Bucket:        aws.String(c.bucket),
			Key:           aws.String(key),
			UploadId:      uploadID,
			PartNumber:    aws.Int32(partNumber),
			Body:          bytes.NewReader(payload),
			ContentLength: aws.Int64(int64(len(payload))),
		})
		if err != nil {
			return err
		}

		parts = append(parts, awss3types.CompletedPart{
			ETag:       partOut.ETag,
			PartNumber: aws.Int32(partNumber),
		})
	}

	if len(parts) == 0 {
		return fmt.Errorf("empty upload body")
	}

	_, err = c.s3.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &awss3types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	return err
}

func readNextUploadChunk(reader io.Reader, maxSize int64) ([]byte, error) {
	buf := make([]byte, int(maxSize))
	n, err := io.ReadFull(reader, buf)
	switch err {
	case nil:
		return buf[:n], nil
	case io.EOF:
		if n == 0 {
			return nil, nil
		}
		return buf[:n], nil
	case io.ErrUnexpectedEOF:
		return buf[:n], nil
	default:
		return nil, err
	}
}

// StatObject performs a HEAD request to retrieve object size and content type
func (c *Client) StatObject(ctx context.Context, key string) (int64, *string, error) {
	out, err := c.s3.HeadObject(ctx, &awss3.HeadObjectInput{
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
		out, err := c.s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
			Bucket:            aws.String(c.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
			MaxKeys:           aws.Int32(1000),
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

func normalizeEndpoint(endpoint string, useSSL bool) string {
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(ep), "http://") || strings.HasPrefix(strings.ToLower(ep), "https://") {
		return ep
	}
	if useSSL {
		return "https://" + ep
	}
	return "http://" + ep
}
