package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	AttachmentBucket = "media"
)

type Client struct {
	c *minio.Client
}

func NewClient(endpoint, accessKeyId, secretAccessKey string, useSSL bool) (*Client, error) {
	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	if !c.IsOnline() {
		return nil, errors.New("minio connection not online")
	}
	return &Client{c: c}, nil
}

// TODO: probably need to add Content-MD5 header, if it has effect
func (c *Client) MakeUploadAttachment(ctx context.Context, channelId, objectId, fileSize int64, objectName string) (string, error) {
	purl, err := c.c.PresignHeader(
		ctx,
		http.MethodPut,
		AttachmentBucket,
		fmt.Sprintf("%d/%d/%s", channelId, objectId, objectName),
		time.Minute,
		url.Values{
			"X-Amz-Meta-ChannelID": []string{fmt.Sprintf("%d", channelId)},
			"X-Amz-Meta-ObjectID":  []string{fmt.Sprintf("%d", objectId)},
		},
		http.Header{
			"Content-Length": []string{fmt.Sprintf("%d", fileSize)},
		})
	if err != nil {
		return "", err
	}

	return purl.String(), nil
}

func (c *Client) RemoveAttachment(ctx context.Context, key, bucket string) error {
	return c.c.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{
		ForceDelete: true,
	})
}
