package applicationcommands

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	CommandTypeChatInput CommandType = 1
	CommandTypeUser      CommandType = 2
	CommandTypeMessage   CommandType = 3
)

const (
	OptionTypeSubCommand      OptionType = 1
	OptionTypeSubCommandGroup OptionType = 2
	OptionTypeString          OptionType = 3
	OptionTypeInteger         OptionType = 4
	OptionTypeBoolean         OptionType = 5
	OptionTypeUser            OptionType = 6
	OptionTypeChannel         OptionType = 7
	OptionTypeRole            OptionType = 8
	OptionTypeMentionable     OptionType = 9
	OptionTypeNumber          OptionType = 10
	OptionTypeAttachment      OptionType = 11
)

const (
	IntegrationTypeGuildInstall IntegrationType = 0
	IntegrationTypeUserInstall  IntegrationType = 1
)

const (
	InteractionContextGuild          InteractionContextType = 0
	InteractionContextBotDM          InteractionContextType = 1
	InteractionContextPrivateChannel InteractionContextType = 2
)

const (
	InteractionTypePing                   InteractionType = 1
	InteractionTypeApplicationCommand     InteractionType = 2
	InteractionTypeMessageComponent       InteractionType = 3
	InteractionTypeAutocomplete           InteractionType = 4
	InteractionTypeModalSubmit            InteractionType = 5
	InteractionTypeApplicationCommandForm InteractionType = 100
)

const (
	ResponseTypePong                         InteractionResponseType = 1
	ResponseTypeChannelMessageWithSource     InteractionResponseType = 4
	ResponseTypeDeferredChannelMessageSource InteractionResponseType = 5
	ResponseTypeDeferredMessageUpdate        InteractionResponseType = 6
	ResponseTypeUpdateMessage                InteractionResponseType = 7
	ResponseTypeAutocompleteResult           InteractionResponseType = 8
	ResponseTypeModal                        InteractionResponseType = 9
)

const (
	MessageFlagEphemeral = 1 << 6
	MessageFlagLoading   = 1 << 7
)

const (
	InteractionDeadline = 3 * time.Second
	InteractionTokenTTL = 15 * time.Minute
)

type CommandType int
type OptionType int
type IntegrationType int
type InteractionContextType int
type InteractionType int
type InteractionResponseType int

type ApplicationCommand struct {
	ID                       int64                      `json:"id,omitempty"`
	ApplicationID            int64                      `json:"application_id,omitempty"`
	GuildID                  *int64                     `json:"guild_id,omitempty"`
	Version                  int64                      `json:"version,omitempty"`
	Type                     CommandType                `json:"type"`
	Name                     string                     `json:"name"`
	NameLocalizations        map[string]string          `json:"name_localizations,omitempty"`
	Description              string                     `json:"description,omitempty"`
	DescriptionLocalizations map[string]string          `json:"description_localizations,omitempty"`
	Options                  []ApplicationCommandOption `json:"options,omitempty"`
	DefaultMemberPermissions *int64                     `json:"default_member_permissions,omitempty"`
	Contexts                 []InteractionContextType   `json:"contexts,omitempty"`
	IntegrationTypes         []IntegrationType          `json:"integration_types,omitempty"`
	NSFW                     bool                       `json:"nsfw,omitempty"`
	BotName                  string                     `json:"bot_name,omitempty"`
	CreatedAt                *time.Time                 `json:"created_at,omitempty"`
	UpdatedAt                *time.Time                 `json:"updated_at,omitempty"`
}

type ApplicationCommandOption struct {
	Type                     OptionType                 `json:"type"`
	Name                     string                     `json:"name"`
	NameLocalizations        map[string]string          `json:"name_localizations,omitempty"`
	Description              string                     `json:"description,omitempty"`
	DescriptionLocalizations map[string]string          `json:"description_localizations,omitempty"`
	Required                 bool                       `json:"required,omitempty"`
	Choices                  []ApplicationCommandChoice `json:"choices,omitempty"`
	Options                  []ApplicationCommandOption `json:"options,omitempty"`
	ChannelTypes             []int                      `json:"channel_types,omitempty"`
	MinValue                 *float64                   `json:"min_value,omitempty"`
	MaxValue                 *float64                   `json:"max_value,omitempty"`
	MinLength                *int                       `json:"min_length,omitempty"`
	MaxLength                *int                       `json:"max_length,omitempty"`
	Autocomplete             bool                       `json:"autocomplete,omitempty"`
}

type ApplicationCommandChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             any               `json:"value"`
}

type ApplicationCommandInteractionData struct {
	ID       int64                                     `json:"id"`
	Name     string                                    `json:"name"`
	Type     CommandType                               `json:"type"`
	Options  []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Resolved *InteractionResolvedData                  `json:"resolved,omitempty"`
	TargetID *int64                                    `json:"target_id,omitempty"`
	GuildID  *int64                                    `json:"guild_id,omitempty"`
}

type ApplicationCommandInteractionDataOption struct {
	Name    string                                    `json:"name"`
	Type    OptionType                                `json:"type"`
	Value   any                                       `json:"value,omitempty"`
	Options []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Focused bool                                      `json:"focused,omitempty"`
}

type InteractionResolvedData struct {
	Users       map[int64]any `json:"users,omitempty"`
	Members     map[int64]any `json:"members,omitempty"`
	Roles       map[int64]any `json:"roles,omitempty"`
	Channels    map[int64]any `json:"channels,omitempty"`
	Messages    map[int64]any `json:"messages,omitempty"`
	Attachments map[int64]any `json:"attachments,omitempty"`
}

type Interaction struct {
	ID                           int64                              `json:"id"`
	ApplicationID                int64                              `json:"application_id"`
	Type                         InteractionType                    `json:"type"`
	Data                         *ApplicationCommandInteractionData `json:"data,omitempty"`
	GuildID                      *int64                             `json:"guild_id,omitempty"`
	ChannelID                    *int64                             `json:"channel_id,omitempty"`
	Member                       any                                `json:"member,omitempty"`
	User                         any                                `json:"user,omitempty"`
	AppPermissions               string                             `json:"app_permissions,omitempty"`
	Locale                       string                             `json:"locale,omitempty"`
	GuildLocale                  string                             `json:"guild_locale,omitempty"`
	Context                      *InteractionContextType            `json:"context,omitempty"`
	Token                        string                             `json:"token"`
	Version                      int                                `json:"version"`
	AuthorizingIntegrationOwners map[string]string                  `json:"authorizing_integration_owners,omitempty"`
}

type InteractionResponse struct {
	Type InteractionResponseType  `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

type InteractionResponseData struct {
	Content     string                     `json:"content,omitempty"`
	Embeds      []any                      `json:"embeds,omitempty"`
	Flags       int                        `json:"flags,omitempty"`
	Choices     []ApplicationCommandChoice `json:"choices,omitempty"`
	CustomID    string                     `json:"custom_id,omitempty"`
	Title       string                     `json:"title,omitempty"`
	Components  []any                      `json:"components,omitempty"`
	Attachments []int64                    `json:"attachments,omitempty"`
}

type InvokeRequest struct {
	Type      InteractionType                           `json:"type"`
	CommandID int64                                     `json:"command_id"`
	GuildID   *int64                                    `json:"guild_id,omitempty"`
	ChannelID int64                                     `json:"channel_id"`
	Options   []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	TargetID  *int64                                    `json:"target_id,omitempty"`
	Locale    string                                    `json:"locale,omitempty"`
}

type InvokeResponse struct {
	InteractionID int64                      `json:"interaction_id"`
	State         string                     `json:"state"`
	Message       any                        `json:"message,omitempty"`
	Ephemeral     *InteractionResponseData   `json:"ephemeral,omitempty"`
	Choices       []ApplicationCommandChoice `json:"choices,omitempty"`
	Error         string                     `json:"error,omitempty"`
}

type InteractionRecord struct {
	ID                int64
	ApplicationID     int64
	CommandID         int64
	Type              InteractionType
	GuildID           *int64
	ChannelID         int64
	InvokerUserID     int64
	TokenHash         string
	TokenPrefix       string
	AppPermissions    int64
	Locale            string
	GuildLocale       string
	Context           InteractionContextType
	AckState          string
	InitialResponseID *int64
	ExpiresAt         time.Time
	CreatedAt         time.Time
	RespondedAt       *time.Time
}

type InteractionPayloadRecord struct {
	ID            int64
	ApplicationID int64
	CommandID     int64
	Type          InteractionType
	GuildID       *int64
	ChannelID     int64
	InvokerUserID int64
	DataJSON      string
	CreatedAt     time.Time
}

type MessageInteractionMetadata struct {
	ID            int64  `json:"id"`
	ApplicationID int64  `json:"application_id"`
	CommandID     int64  `json:"command_id"`
	CommandName   string `json:"command_name"`
	UserID        int64  `json:"user_id"`
}

const (
	AckStatePending   = "pending"
	AckStateDeferred  = "deferred"
	AckStateResponded = "responded"
	AckStateExpired   = "expired"
)

var (
	commandNameRe = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)
	localeRe      = regexp.MustCompile(`^[a-z]{2}(?:-[A-Z]{2})?$`)
)

func NormalizeCommand(cmd ApplicationCommand) ApplicationCommand {
	cmd.Name = strings.ToLower(strings.TrimSpace(cmd.Name))
	cmd.Description = strings.TrimSpace(cmd.Description)
	if cmd.Type == 0 {
		cmd.Type = CommandTypeChatInput
	}
	if cmd.Contexts == nil {
		cmd.Contexts = []InteractionContextType{InteractionContextGuild, InteractionContextBotDM}
	}
	if cmd.IntegrationTypes == nil {
		cmd.IntegrationTypes = []IntegrationType{IntegrationTypeGuildInstall}
	}
	cmd.Options = normalizeOptions(cmd.Options)
	return cmd
}

func ValidateCommand(cmd ApplicationCommand) error {
	cmd = NormalizeCommand(cmd)
	if !validCommandType(cmd.Type) {
		return fmt.Errorf("unsupported command type %d", cmd.Type)
	}
	if !commandNameRe.MatchString(cmd.Name) {
		return errors.New("command name must be 1-32 lowercase letters, numbers, underscores, or hyphens")
	}
	if err := validateLocalizations("name_localizations", cmd.NameLocalizations, true); err != nil {
		return err
	}
	if err := validateContexts(cmd.Contexts); err != nil {
		return err
	}
	if err := validateIntegrationTypes(cmd.IntegrationTypes); err != nil {
		return err
	}
	switch cmd.Type {
	case CommandTypeChatInput:
		if l := len([]rune(cmd.Description)); l < 1 || l > 100 {
			return errors.New("chat input command description must be 1-100 characters")
		}
		if len(cmd.Options) > 25 {
			return errors.New("chat input command cannot have more than 25 options")
		}
		if err := validateLocalizations("description_localizations", cmd.DescriptionLocalizations, false); err != nil {
			return err
		}
		return validateOptions(cmd.Options, optionDepthRoot)
	case CommandTypeUser, CommandTypeMessage:
		if strings.TrimSpace(cmd.Description) != "" || len(cmd.Options) > 0 {
			return errors.New("context menu commands cannot have descriptions or options")
		}
		return nil
	default:
		return nil
	}
}

func MarshalJSONField(value any) (string, error) {
	if value == nil {
		return "null", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func UnmarshalJSONField[T any](raw string, fallback T) T {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return fallback
	}
	return out
}

func CommandDataJSON(cmd ApplicationCommandInteractionData) string {
	raw, err := json.Marshal(cmd)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func ParseCommandData(raw string) *ApplicationCommandInteractionData {
	if raw == "" {
		return nil
	}
	var data ApplicationCommandInteractionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	return &data
}

func HasResponseFlag(data *InteractionResponseData, flag int) bool {
	return data != nil && data.Flags&flag == flag
}

func normalizeOptions(options []ApplicationCommandOption) []ApplicationCommandOption {
	if len(options) == 0 {
		return nil
	}
	out := make([]ApplicationCommandOption, 0, len(options))
	for _, option := range options {
		option.Name = strings.ToLower(strings.TrimSpace(option.Name))
		option.Description = strings.TrimSpace(option.Description)
		option.Options = normalizeOptions(option.Options)
		out = append(out, option)
	}
	return out
}

func validCommandType(t CommandType) bool {
	return t == CommandTypeChatInput || t == CommandTypeUser || t == CommandTypeMessage
}

func validateContexts(values []InteractionContextType) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("contexts must include 1-3 values")
	}
	seen := map[InteractionContextType]struct{}{}
	for _, value := range values {
		switch value {
		case InteractionContextGuild, InteractionContextBotDM, InteractionContextPrivateChannel:
		default:
			return fmt.Errorf("unsupported context %d", value)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("duplicate context %d", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateIntegrationTypes(values []IntegrationType) error {
	if len(values) == 0 || len(values) > 2 {
		return errors.New("integration_types must include 1-2 values")
	}
	seen := map[IntegrationType]struct{}{}
	for _, value := range values {
		switch value {
		case IntegrationTypeGuildInstall, IntegrationTypeUserInstall:
		default:
			return fmt.Errorf("unsupported integration type %d", value)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("duplicate integration type %d", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateLocalizations(field string, values map[string]string, name bool) error {
	for locale, value := range values {
		if !localeRe.MatchString(locale) {
			return fmt.Errorf("%s contains invalid locale %q", field, locale)
		}
		length := len([]rune(strings.TrimSpace(value)))
		if name {
			if length < 1 || length > 32 || !commandNameRe.MatchString(strings.ToLower(value)) {
				return fmt.Errorf("%s[%s] must be a valid 1-32 character command name", field, locale)
			}
		} else if length < 1 || length > 100 {
			return fmt.Errorf("%s[%s] must be 1-100 characters", field, locale)
		}
	}
	return nil
}

type optionDepth int

const (
	optionDepthRoot optionDepth = iota
	optionDepthGroup
	optionDepthSubCommand
)

func validateOptions(options []ApplicationCommandOption, depth optionDepth) error {
	optionalSeen := false
	names := make(map[string]struct{}, len(options))
	for _, option := range options {
		if !validOptionType(option.Type) {
			return fmt.Errorf("unsupported option type %d", option.Type)
		}
		if !commandNameRe.MatchString(option.Name) {
			return fmt.Errorf("option %q name must be 1-32 lowercase letters, numbers, underscores, or hyphens", option.Name)
		}
		if _, ok := names[option.Name]; ok {
			return fmt.Errorf("duplicate option name %q", option.Name)
		}
		names[option.Name] = struct{}{}
		if len([]rune(option.Description)) < 1 || len([]rune(option.Description)) > 100 {
			return fmt.Errorf("option %q description must be 1-100 characters", option.Name)
		}
		if err := validateLocalizations("option.name_localizations", option.NameLocalizations, true); err != nil {
			return err
		}
		if err := validateLocalizations("option.description_localizations", option.DescriptionLocalizations, false); err != nil {
			return err
		}
		if option.Type == OptionTypeSubCommandGroup {
			if depth != optionDepthRoot {
				return errors.New("subcommand groups are only allowed at the root level")
			}
			if option.Required {
				return fmt.Errorf("subcommand group %q cannot be required", option.Name)
			}
			if len(option.Options) == 0 || len(option.Options) > 25 {
				return fmt.Errorf("subcommand group %q must contain 1-25 subcommands", option.Name)
			}
			for _, nested := range option.Options {
				if nested.Type != OptionTypeSubCommand {
					return fmt.Errorf("subcommand group %q can only contain subcommands", option.Name)
				}
			}
			if err := validateOptions(option.Options, optionDepthGroup); err != nil {
				return err
			}
			continue
		}
		if option.Type == OptionTypeSubCommand {
			if depth == optionDepthSubCommand {
				return fmt.Errorf("subcommand %q cannot contain another subcommand", option.Name)
			}
			if option.Required {
				return fmt.Errorf("subcommand %q cannot be required", option.Name)
			}
			if len(option.Options) > 25 {
				return fmt.Errorf("subcommand %q cannot have more than 25 options", option.Name)
			}
			if err := validateOptions(option.Options, optionDepthSubCommand); err != nil {
				return err
			}
			continue
		}
		if depth == optionDepthGroup {
			return fmt.Errorf("subcommand group can only contain subcommands; got %q", option.Name)
		}
		if optionalSeen && option.Required {
			return fmt.Errorf("required option %q must be listed before optional options", option.Name)
		}
		if !option.Required {
			optionalSeen = true
		}
		if len(option.Options) > 0 {
			return fmt.Errorf("leaf option %q cannot contain nested options", option.Name)
		}
		if len(option.Choices) > 25 {
			return fmt.Errorf("option %q cannot have more than 25 choices", option.Name)
		}
		if option.Autocomplete && len(option.Choices) > 0 {
			return fmt.Errorf("option %q cannot define choices and autocomplete", option.Name)
		}
		if err := validateLeafOption(option); err != nil {
			return err
		}
	}
	return nil
}

func validOptionType(t OptionType) bool {
	return t >= OptionTypeSubCommand && t <= OptionTypeAttachment
}

func validateLeafOption(option ApplicationCommandOption) error {
	if len(option.ChannelTypes) > 0 && option.Type != OptionTypeChannel {
		return fmt.Errorf("channel_types is only valid for channel option %q", option.Name)
	}
	if (option.MinLength != nil || option.MaxLength != nil) && option.Type != OptionTypeString {
		return fmt.Errorf("min_length/max_length are only valid for string option %q", option.Name)
	}
	if option.MinLength != nil && (*option.MinLength < 0 || *option.MinLength > 6000) {
		return fmt.Errorf("option %q min_length must be between 0 and 6000", option.Name)
	}
	if option.MaxLength != nil && (*option.MaxLength < 1 || *option.MaxLength > 6000) {
		return fmt.Errorf("option %q max_length must be between 1 and 6000", option.Name)
	}
	if option.MinLength != nil && option.MaxLength != nil && *option.MinLength > *option.MaxLength {
		return fmt.Errorf("option %q min_length cannot exceed max_length", option.Name)
	}
	if (option.MinValue != nil || option.MaxValue != nil) && option.Type != OptionTypeInteger && option.Type != OptionTypeNumber {
		return fmt.Errorf("min_value/max_value are only valid for numeric option %q", option.Name)
	}
	if option.MinValue != nil && option.MaxValue != nil && *option.MinValue > *option.MaxValue {
		return fmt.Errorf("option %q min_value cannot exceed max_value", option.Name)
	}
	for _, choice := range option.Choices {
		if len([]rune(strings.TrimSpace(choice.Name))) < 1 || len([]rune(choice.Name)) > 100 {
			return fmt.Errorf("choice for option %q must have a 1-100 character name", option.Name)
		}
		if choice.Value == nil {
			return fmt.Errorf("choice %q for option %q requires a value", choice.Name, option.Name)
		}
	}
	return nil
}
