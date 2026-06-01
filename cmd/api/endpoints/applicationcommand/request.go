package applicationcommand

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
)

type snowflakeID int64

func (s *snowflakeID) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*s = 0
		return nil
	}
	if strings.HasPrefix(raw, `"`) {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			*s = 0
			return nil
		}
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid snowflake %q", text)
		}
		*s = snowflakeID(id)
		return nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid snowflake %q", raw)
	}
	*s = snowflakeID(id)
	return nil
}

func (s snowflakeID) int64() int64 {
	return int64(s)
}

func snowflakePtr(s *snowflakeID) *int64 {
	if s == nil || *s == 0 {
		return nil
	}
	id := int64(*s)
	return &id
}

type clientInteractionRequest struct {
	Type          appcmd.InteractionType                           `json:"type"`
	ApplicationID snowflakeID                                      `json:"application_id"`
	GuildID       *snowflakeID                                     `json:"guild_id"`
	ChannelID     snowflakeID                                      `json:"channel_id"`
	CommandID     snowflakeID                                      `json:"command_id"`
	Options       []appcmd.ApplicationCommandInteractionDataOption `json:"options"`
	TargetID      *snowflakeID                                     `json:"target_id"`
	Locale        string                                           `json:"locale"`
	Data          *clientInteractionData                           `json:"data"`
}

type clientInteractionData struct {
	ID       snowflakeID                                      `json:"id"`
	Name     string                                           `json:"name"`
	Type     appcmd.CommandType                               `json:"type"`
	Options  []appcmd.ApplicationCommandInteractionDataOption `json:"options"`
	TargetID *snowflakeID                                     `json:"target_id"`
}

func parseInvokeRequest(body []byte, fallbackType appcmd.InteractionType) (appcmd.InvokeRequest, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return appcmd.InvokeRequest{}, fmt.Errorf("empty interaction body")
	}
	var raw clientInteractionRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return appcmd.InvokeRequest{}, err
	}
	interactionType := raw.Type
	if interactionType == 0 {
		interactionType = fallbackType
	}
	commandID := raw.CommandID.int64()
	options := raw.Options
	targetID := snowflakePtr(raw.TargetID)
	if raw.Data != nil {
		if commandID == 0 {
			commandID = raw.Data.ID.int64()
		}
		if options == nil {
			options = raw.Data.Options
		}
		if targetID == nil {
			targetID = snowflakePtr(raw.Data.TargetID)
		}
	}
	return appcmd.InvokeRequest{
		Type:      interactionType,
		CommandID: commandID,
		GuildID:   snowflakePtr(raw.GuildID),
		ChannelID: raw.ChannelID.int64(),
		Options:   options,
		TargetID:  targetID,
		Locale:    raw.Locale,
	}, nil
}
