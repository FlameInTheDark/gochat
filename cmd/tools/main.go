package main

import (
	"context"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "gctools",
		Usage: "Utility tools for GoChat services",
		Commands: []*cli.Command{
			permissions(),
			tokens(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
