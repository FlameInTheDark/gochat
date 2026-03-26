package server

import (
	"log/slog"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

type Server struct {
	app   *fiber.App
	cache *kvs.Cache
}

func NewServer(prefork ...bool) *Server {
	pf := len(prefork) > 0 && prefork[0]
	app := fiber.New(fiber.Config{
		Prefork:               pf,
		DisableStartupMessage: true,
		BodyLimit:             100 * 1024 * 1024, // 100MB
		ReadBufferSize:        8192,
		WriteBufferSize:       8192,
	})
	rc := recover.ConfigDefault
	rc.EnableStackTrace = true
	app.Get("/healthz", healthzHandler)
	app.Use(observability.RequestContextMiddleware())
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
	s.app.Use(observability.RequestLogger(logger))
}

func (s *Server) WithLoggerLevel(logger *slog.Logger, level slog.Level) {
	s.app.Use(observability.RequestLoggerWithLevel(logger, level))
}

func (s *Server) WithCORS() {
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
}

func (s *Server) WithCompression() {
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
}

func (s *Server) WithMetrics(serviceName ...string) {
	name := "gochat-api"
	if len(serviceName) > 0 && serviceName[0] != "" {
		name = serviceName[0]
	}
	s.app.Use(observability.NewHTTPServerTelemetry(name).Middleware())
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
