package server

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	slogfiber "github.com/samber/slog-fiber"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
)

type Server struct {
	app   *fiber.App
	cache *kvs.Cache
}

func NewServer() *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	rc := recover.ConfigDefault
	rc.EnableStackTrace = true
	app.Get("/healthz", healthzHandler)
	app.Use(recover.New(rc))
	return &Server{app: app}
}

func healthzHandler(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (s *Server) Register(base string, components ...Entity) {
	group := s.app.Group
	if base != "" {
		group = s.app.Group(base).Group
	}
	for _, c := range components {
		c.Init(group(c.Name()))
	}
}

func (s *Server) WithLogger(logger *slog.Logger) {
	logMiddleware := slogfiber.NewWithFilters(
		logger,
		slogfiber.IgnorePath("/metrics"),
	)
	s.app.Use(logMiddleware)
}

func (s *Server) WithCORS() {
	// Initialize default config
	s.app.Use(cors.New())

	// Or extend your config for customization
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
}

func (s *Server) WithMetrics(serviceName ...string) {
	name := "gochat-api"
	if len(serviceName) > 0 && serviceName[0] != "" {
		name = serviceName[0]
	}

	// Expose default Prometheus registry at /metrics using promhttp
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	s.app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	})

	reqCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gochat",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"service", "method", "code", "route"})

	reqDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gochat",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "method", "code", "route"})

	prometheus.MustRegister(reqCounter, reqDuration)

	s.app.Use(func(c *fiber.Ctx) error {
		// Skip non-business endpoints from custom metrics
		p := c.Path()
		if p == "/metrics" || p == "/healthz" {
			return c.Next()
		}
		start := time.Now()
		err := c.Next()
		code := c.Response().StatusCode()
		method := c.Method()
		route := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			route = r.Path
		}
		reqCounter.WithLabelValues(name, method, strconv.Itoa(code), route).Inc()
		reqDuration.WithLabelValues(name, method, strconv.Itoa(code), route).Observe(time.Since(start).Seconds())
		return err
	})
}

func (s *Server) WithCache(c *kvs.Cache) {
	s.cache = c
}

func (s *Server) WithSwagger(app string) {
	s.app.Use(swagger.New(swagger.Config{
		BasePath: "/docs/",
		FilePath: "./docs/" + app + "/swagger.json",
		Path:     "swagger",
		Title:    "GoChat API",
	}))
}

func (s *Server) Start(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Close() error {
	return s.app.Shutdown()
}

func (s *Server) Use(args ...interface{}) {
	s.app.Use(args...)
}
