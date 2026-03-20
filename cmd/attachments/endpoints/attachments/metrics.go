package attachments

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	attachmentsBytesTransferred metric.Int64Counter
)

func init() {
	attachmentsBytesTransferred, _ = otel.Meter("gochat-attachments").Int64Counter("gochat.attachments.bytes_transferred")
}

func incTransferred(kind string, n int64) {
	if n <= 0 {
		return
	}
	attachmentsBytesTransferred.Add(
		context.Background(),
		n,
		metric.WithAttributes(attribute.String("kind", kind)),
	)
}
