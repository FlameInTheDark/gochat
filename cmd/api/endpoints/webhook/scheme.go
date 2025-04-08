package webhook

import (
	"time"
)

const (
	ErrUnableToParseRequestBody = "unable to parse request body"
	ErrNoEventsProvided         = "no events provided"
	ErrUnableToExtractID        = "unable to extract ID"
	ErrAttachmentNotFount       = "attachment not found"
	ErrUnableToDoneAttachment   = "unable to done attachment"
	ErrUnableToRemoveAttachment = "unable to remove attachment"
	ErrIncorrectFileSize        = "incorrect file size"
)

const (
	S3EventPut    = "s3:ObjectCreated:Put"
	S3EventDelete = "s3:ObjectRemoved:Delete"
)

type S3Event struct {
	EventName string          `json:"EventName"`
	Key       string          `json:"Key"`
	Records   []S3EventRecord `json:"Records"`
}

type S3EventRecord struct {
	EventVersion      string              `json:"eventVersion"`
	EventSource       string              `json:"eventSource"`
	AWSRegion         string              `json:"awsRegion"`
	EventTime         time.Time           `json:"eventTime"`
	EventName         string              `json:"eventName"`
	UserIdentity      S3Identity          `json:"userIdentity"`
	RequestParameters S3RequestParameters `json:"requestParameters"`
	ResponseElements  map[string]string   `json:"responseElements"`
	S3                S3Element           `json:"s3"`
	Source            S3Source            `json:"source"`
}

type S3Identity struct {
	PrincipalID string `json:"principalId"`
}

type S3RequestParameters struct {
	PrincipalID     string `json:"principalId"`
	Region          string `json:"region"`
	SourceIPAddress string `json:"sourceIPAddress"`
}

type S3Element struct {
	S3SchemaVersion string   `json:"s3SchemaVersion"`
	ConfigurationID string   `json:"configurationId"`
	Bucket          S3Bucket `json:"bucket"`
	Object          S3Object `json:"object"`
}

type S3Bucket struct {
	Name          string     `json:"name"`
	OwnerIdentity S3Identity `json:"ownerIdentity"`
	ARN           string     `json:"arn"`
}

type S3Object struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"eTag"`
	ContentType  *string           `json:"contentType"`
	UserMetadata map[string]string `json:"userMetadata"`
	Sequencer    string            `json:"sequencer"`
}

type S3Source struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}
