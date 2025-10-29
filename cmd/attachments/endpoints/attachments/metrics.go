package attachments

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	attachmentsBytesTransferred = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gochat",
			Subsystem: "attachments",
			Name:      "bytes_transferred_total",
			Help:      "Total bytes successfully transferred for attachments (original objects)",
		},
		[]string{"kind"}, // kind: image|video|other
	)
)

func init() {
	prometheus.MustRegister(attachmentsBytesTransferred)
}

func incTransferred(kind string, n int64) {
	if n <= 0 {
		return
	}
	attachmentsBytesTransferred.WithLabelValues(kind).Add(float64(n))
}
