package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	cli "github.com/urfave/cli/v3"

	perm "github.com/FlameInTheDark/gochat/internal/permissions"
)

func main() {
	app := &cli.Command{
		Name:  "gochat-tools",
		Usage: "Utility tools for GoChat services",
		Commands: []*cli.Command{
			// ---- permissions decoder ----
			{
				Name:    "perms-decode",
				Aliases: []string{"perms", "perm"},
				Usage:   "Decode an int64 permission bitmask into permission names",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "value", Aliases: []string{"v"}, Usage: "permission bitmask (supports 123, 0x..., 0b...)", Required: true},
					&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "text", Usage: "output: text|json"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					raw := strings.TrimSpace(cmd.String("value"))
					val, err := parseInt64Flexible(raw)
					if err != nil {
						return fmt.Errorf("invalid value %q: %w", raw, err)
					}
					names := decodePermissions(val)
					switch cmd.String("format") {
					case "json":
						out := struct {
							Value int64    `json:"value"`
							Hex   string   `json:"hex"`
							Names []string `json:"names"`
						}{Value: val, Hex: fmt.Sprintf("0x%016x", uint64(val)), Names: names}
						enc := json.NewEncoder(os.Stdout)
						enc.SetIndent("", "  ")
						return enc.Encode(out)
					default:
						fmt.Printf("value: %d (%s)\n", val, fmt.Sprintf("0x%016x", uint64(val)))
						for _, n := range names {
							fmt.Println(n)
						}
						return nil
					}
				},
			},
			{
				Name:  "gen-token",
				Usage: "Generate webhook JWT token and id for a service (works with webhook auth)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Usage: "service type: sfu|attachments", Required: true},
					&cli.StringFlag{Name: "id", Aliases: []string{"i"}, Usage: "service id (UUIDv4). If empty, a new UUID is generated."},
					&cli.StringFlag{Name: "secret", Aliases: []string{"s"}, Usage: "HS256 secret (webhook jwt_secret)", Required: true},
					&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "text", Usage: "output: text|json"},
					&cli.BoolFlag{Name: "header", Usage: "print X-Webhook-Token header"},
					&cli.BoolFlag{Name: "curl", Usage: "print curl example"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					typ := cmd.String("type")
					switch typ {
					case "sfu", "attachments":
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
							fmt.Printf("curl -X POST 'http://webhook:3200/webhook/sfu/heartbeat' \\\n+  -H 'Content-Type: application/json' \\\n+  -H 'X-Webhook-Token: %s' \\\n+  -d '{\"id\":\"%s\",\"region\":\"eu\",\"url\":\"wss://sfu.example.com/sfu/signal\",\"load\":0}'\n", signed, id)
						case "attachments":
							fmt.Printf("curl -X POST 'http://webhook:3200/webhook/attachments/finalize' \\\n+  -H 'Content-Type: application/json' \\\n+  -H 'X-Webhook-Token: %s' \\\n+  -d '{\"id\":2230469276416868352,\"channel_id\":2230469276416868352,\"url\":\"https://cdn.example.com/file\"}'\n", signed)
						}
					}
					return nil
				},
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// parseInt64Flexible parses an int64 from decimal (default), hex (0x/0X), or binary (0b/0B).
func parseInt64Flexible(s string) (int64, error) {
	ls := strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(ls, "0x") {
		u, err := strconv.ParseUint(ls[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return int64(u), nil
	}
	if strings.HasPrefix(ls, "-0x") { // handle negative hex
		u, err := strconv.ParseUint(ls[3:], 16, 64)
		if err != nil {
			return 0, err
		}
		return -int64(u), nil
	}
	if strings.HasPrefix(ls, "0b") {
		u, err := strconv.ParseUint(ls[2:], 2, 64)
		if err != nil {
			return 0, err
		}
		return int64(u), nil
	}
	if strings.HasPrefix(ls, "-0b") {
		u, err := strconv.ParseUint(ls[3:], 2, 64)
		if err != nil {
			return 0, err
		}
		return -int64(u), nil
	}
	return strconv.ParseInt(ls, 10, 64)
}

type permDef struct {
	Name string
	Bit  perm.RolePermission
}

func allPerms() []permDef {
	return []permDef{
		{"PermServerViewChannels", perm.PermServerViewChannels},
		{"PermServerManageChannels", perm.PermServerManageChannels},
		{"PermServerManageRoles", perm.PermServerManageRoles},
		{"PermServerViewAuditLog", perm.PermServerViewAuditLog},
		{"PermServerManage", perm.PermServerManage},
		{"PermMembershipCreateInvite", perm.PermMembershipCreateInvite},
		{"PermMembershipChangeNickname", perm.PermMembershipChangeNickname},
		{"PermMembershipManageNickname", perm.PermMembershipManageNickname},
		{"PermMembershipKickMembers", perm.PermMembershipKickMembers},
		{"PermMembershipBanMembers", perm.PermMembershipBanMembers},
		{"PermMembershipTimeoutMembers", perm.PermMembershipTimeoutMembers},
		{"PermTextSendMessage", perm.PermTextSendMessage},
		{"PermTextSendMessageInThreads", perm.PermTextSendMessageInThreads},
		{"PermTextCreateThreads", perm.PermTextCreateThreads},
		{"PermTextAttachFiles", perm.PermTextAttachFiles},
		{"PermTextAddReactions", perm.PermTextAddReactions},
		{"PermTextMentionRoles", perm.PermTextMentionRoles},
		{"PermTextManageMessages", perm.PermTextManageMessages},
		{"PermTextManageThreads", perm.PermTextManageThreads},
		{"PermTextReadMessageHistory", perm.PermTextReadMessageHistory},
		{"PermVoiceConnect", perm.PermVoiceConnect},
		{"PermVoiceSpeak", perm.PermVoiceSpeak},
		{"PermVoiceVideo", perm.PermVoiceVideo},
		{"PermVoiceMuteMembers", perm.PermVoiceMuteMembers},
		{"PermVoiceDeafenMembers", perm.PermVoiceDeafenMembers},
		{"PermVoiceMoveMembers", perm.PermVoiceMoveMembers},
		{"PermAdministrator", perm.PermAdministrator},
	}
}

func decodePermissions(val int64) []string {
	var names []string
	for _, p := range allPerms() {
		if val&int64(p.Bit) != 0 {
			names = append(names, p.Name)
		}
	}
	return names
}
