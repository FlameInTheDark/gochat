package server

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app *fiber.App
}

func NewServer() *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())
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

func (s *Server) WithLogger() {
	s.app.Use(logger.New())
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
