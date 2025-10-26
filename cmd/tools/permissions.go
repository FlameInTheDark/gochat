package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	perm "github.com/FlameInTheDark/gochat/internal/permissions"
)

func permissions() *cli.Command {
	return &cli.Command{
		Name:    "permissions",
		Aliases: []string{"perm"},
		Usage:   "Permissions operations command",
		Commands: []*cli.Command{
			permissionsDecode(),
		},
	}
}

func permissionsDecode() *cli.Command {
	return &cli.Command{
		Name:  "decode",
		Usage: "Decode an int64 permission bitmask into permission names",
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
