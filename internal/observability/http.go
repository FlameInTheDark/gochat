package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric/noop"
)

func NewHTTPTransport(name string, base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return otelhttp.NewTransport(
		base,
		// We keep outbound HTTP traces, but disable otelhttp's generic client metrics
		// because the project already emits lower-cardinality gochat.dependency.* metrics.
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
			if req == nil {
				return name
			}
			if name == "" {
				return req.Method + " " + req.URL.Host
			}
			return name + " " + req.Method
		}),
	)
}

func NewHTTPClient(base *http.Client, name string) *http.Client {
	if base == nil {
		base = &http.Client{}
	}
	client := *base
	client.Transport = NewHTTPTransport(name, client.Transport)
	return &client
}
