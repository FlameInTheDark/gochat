package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/shut"
)

//	@title			GoChat API
//	@version		1.0
//	@description	This is an API for the GoChat

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:3000
//	@BasePath	/api/v1

//	@securityDefinitions.basic	ApiKeyAuth

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shutter := shut.NewShutter(logger)
	defer shutter.Down()

	app, err := NewApp(shutter, logger)
	if err != nil {
		logger.Error("Unable to create app", slog.String("error", err.Error()))
		return
	}
	app.Start(":3000")
}
