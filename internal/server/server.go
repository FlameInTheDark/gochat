package server

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/cache/vkc"
)

type Server struct {
	app   *fiber.App
	cache *vkc.Cache
}

func NewServer() *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	rc := recover.ConfigDefault
	rc.EnableStackTrace = true
	app.Use(recover.New(rc))
	return &Server{app: app}
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

func (s *Server) WithMetrics() {
	prom := fiberprometheus.New("gochat-api")
	prom.RegisterAt(s.app, "/metrics")
	prom.SetSkipPaths([]string{"/healthz"})
	s.app.Use(prom.Middleware)
}

func (s *Server) WithCache(c *vkc.Cache) {
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
