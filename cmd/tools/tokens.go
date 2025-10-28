package main

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func tokens() *cli.Command {
	return &cli.Command{
		Name:  "tokens",
		Usage: "Tools to operate with tokens",
		Commands: []*cli.Command{
			webhook(),
		},
	}
}

func webhook() *cli.Command {
	return &cli.Command{
		Name:  "webhook",
		Usage: "Tools to operate with webhook auth tokens",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Usage:   "Generate webhook JWT token and id for a service (works with webhook auth)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Usage: "service type: sfu|attachments|prom", Required: true},
					&cli.StringFlag{Name: "id", Aliases: []string{"i"}, Usage: "service id (UUIDv4). If empty, a new UUID is generated."},
					&cli.StringFlag{Name: "secret", Aliases: []string{"s"}, Usage: "HS256 secret (webhook jwt_secret)", Required: true},
					&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "text", Usage: "output: text|json"},
					&cli.BoolFlag{Name: "header", Usage: "print X-Webhook-Token header"},
					&cli.BoolFlag{Name: "curl", Usage: "print curl example"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					typ := cmd.String("type")
					switch typ {
					case "sfu", "attachments", "prom":
					default:
						return fmt.Errorf("unknown type: %s", typ)
					}
					id := cmd.String("id")
					if id == "" {
						id = uuid.NewString()
					}
					secret := cmd.String("secret")

					tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"typ": typ, "id": id})
					signed, err := tok.SignedString([]byte(secret))
					if err != nil {
						return err
					}

					switch cmd.String("format") {
					case "json":
						fmt.Printf("{\n  \"type\": \"%s\",\n  \"id\": \"%s\",\n  \"token\": \"%s\"\n}\n", typ, id, signed)
					default:
						fmt.Printf("type=%s\n", typ)
						fmt.Printf("id=%s\n", id)
						fmt.Printf("token=%s\n", signed)
					}
					if cmd.Bool("header") {
						fmt.Printf("X-Webhook-Token: %s\n", signed)
					}
					if cmd.Bool("curl") {
						switch typ {
						case "sfu":
							fmt.Printf("curl -X POST 'http://example.com/api/v1/webhook/sfu/heartbeat' \\\n+  -H 'Content-Type: application/json' \\\n+  -H 'X-Webhook-Token: %s' \\\n+  -d '{\"id\":\"%s\",\"region\":\"eu\",\"url\":\"wss://sfu.example.com/signal\",\"load\":0}'\n", signed, id)
						case "attachments":
							fmt.Printf("curl -X POST 'http://example.com/api/v1/webhook/attachments/finalize' \\\n+  -H 'Content-Type: application/json' \\\n+  -H 'X-Webhook-Token: %s' \\\n+  -d '{\"id\":2230469276416868352,\"channel_id\":2230469276416868352,\"url\":\"https://cdn.example.com/file\"}'\n", signed)
						}
					}
					return nil
				},
			},
		},
	}
}
