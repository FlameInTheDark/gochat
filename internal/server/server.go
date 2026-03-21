package server

import (
	"log/slog"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

type Server struct {
	app   *fiber.App
	cache *kvs.Cache
}

func NewServer() *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             100 * 1024 * 1024, // 100MB
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
