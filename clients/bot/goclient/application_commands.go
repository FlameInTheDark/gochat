package goclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNilInteraction is returned when an interaction helper receives nil.
var ErrNilInteraction = errors.New("nil interaction")

// ApplicationCommandType identifies a Discord-compatible application command kind.
type ApplicationCommandType int

const (
	ApplicationCommandChatInput ApplicationCommandType = 1
	ApplicationCommandUser      ApplicationCommandType = 2
	ApplicationCommandMessage   ApplicationCommandType = 3
)

// ApplicationCommandOptionType identifies a chat input command option kind.
type ApplicationCommandOptionType int

const (
	ApplicationCommandOptionSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionRole            ApplicationCommandOptionType = 8
	ApplicationCommandOptionMentionable     ApplicationCommandOptionType = 9
	ApplicationCommandOptionNumber          ApplicationCommandOptionType = 10
	ApplicationCommandOptionAttachment      ApplicationCommandOptionType = 11
)

// ApplicationIntegrationType identifies where an application command is installed.
type ApplicationIntegrationType int

const (
	ApplicationIntegrationGuildInstall ApplicationIntegrationType = 0
	ApplicationIntegrationUserInstall  ApplicationIntegrationType = 1
)

// InteractionContextType identifies where an interaction may be used.
type InteractionContextType int

const (
	InteractionContextGuild          InteractionContextType = 0
	InteractionContextBotDM          InteractionContextType = 1
	InteractionContextPrivateChannel InteractionContextType = 2
)

// InteractionType identifies a Discord-compatible interaction payload kind.
type InteractionType int

const (
	InteractionPing               InteractionType = 1
	InteractionApplicationCommand InteractionType = 2
	InteractionMessageComponent   InteractionType = 3
	InteractionAutocomplete       InteractionType = 4
	InteractionModalSubmit        InteractionType = 5
)

// InteractionResponseType identifies how a bot acknowledges an interaction.
type InteractionResponseType int

const (
	InteractionResponsePong                         InteractionResponseType = 1
	InteractionResponseChannelMessageWithSource     InteractionResponseType = 4
	InteractionResponseDeferredChannelMessageSource InteractionResponseType = 5
	InteractionResponseDeferredMessageUpdate        InteractionResponseType = 6
	InteractionResponseUpdateMessage                InteractionResponseType = 7
	InteractionApplicationCommandAutocompleteResult InteractionResponseType = 8
	InteractionResponseModal                        InteractionResponseType = 9
)

const (
	// MessageFlagsEphemeral makes an interaction response visible only to the invoker.
	MessageFlagsEphemeral = 1 << 6
	// MessageFlagsLoading marks a deferred interaction response.
	MessageFlagsLoading = 1 << 7
)

// ApplicationCommand is the bot-visible command definition.
type ApplicationCommand struct {
	ID                       int64                        `json:"id,omitempty"`
	ApplicationID            int64                        `json:"application_id,omitempty"`
	GuildID                  *int64                       `json:"guild_id,omitempty"`
	Version                  int64                        `json:"version,omitempty"`
	Type                     ApplicationCommandType       `json:"type"`
	Name                     string                       `json:"name"`
	NameLocalizations        map[string]string            `json:"name_localizations,omitempty"`
	Description              string                       `json:"description,omitempty"`
	DescriptionLocalizations map[string]string            `json:"description_localizations,omitempty"`
	Options                  []*ApplicationCommandOption  `json:"options,omitempty"`
	DefaultMemberPermissions *int64                       `json:"default_member_permissions,omitempty"`
	Contexts                 []InteractionContextType     `json:"contexts,omitempty"`
	IntegrationTypes         []ApplicationIntegrationType `json:"integration_types,omitempty"`
	NSFW                     bool                         `json:"nsfw,omitempty"`
	BotName                  string                       `json:"bot_name,omitempty"`
}

// ApplicationCommandOption describes a chat input command option.
type ApplicationCommandOption struct {
	Type                     ApplicationCommandOptionType      `json:"type"`
	Name                     string                            `json:"name"`
	NameLocalizations        map[string]string                 `json:"name_localizations,omitempty"`
	Description              string                            `json:"description,omitempty"`
	DescriptionLocalizations map[string]string                 `json:"description_localizations,omitempty"`
	Required                 bool                              `json:"required,omitempty"`
	Choices                  []*ApplicationCommandOptionChoice `json:"choices,omitempty"`
	Options                  []*ApplicationCommandOption       `json:"options,omitempty"`
	ChannelTypes             []int                             `json:"channel_types,omitempty"`
	MinValue                 *float64                          `json:"min_value,omitempty"`
	MaxValue                 *float64                          `json:"max_value,omitempty"`
	MinLength                *int                              `json:"min_length,omitempty"`
	MaxLength                *int                              `json:"max_length,omitempty"`
	Autocomplete             bool                              `json:"autocomplete,omitempty"`
}

// ApplicationCommandOptionChoice describes a fixed value for a command option.
type ApplicationCommandOptionChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             any               `json:"value"`
}

// ApplicationCommandInteractionData is the parsed command invocation payload.
type ApplicationCommandInteractionData struct {
	ID       int64                                      `json:"id"`
	Name     string                                     `json:"name"`
	Type     ApplicationCommandType                     `json:"type"`
	Options  []*ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Resolved *InteractionResolvedData                   `json:"resolved,omitempty"`
	TargetID *int64                                     `json:"target_id,omitempty"`
	GuildID  *int64                                     `json:"guild_id,omitempty"`
}

// ApplicationCommandInteractionDataOption is an invoked option or subcommand node.
type ApplicationCommandInteractionDataOption struct {
	Name    string                                     `json:"name"`
	Type    ApplicationCommandOptionType               `json:"type"`
	Value   any                                        `json:"value,omitempty"`
	Options []*ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Focused bool                                       `json:"focused,omitempty"`
}

// StringValue returns an option value as a string when possible.
func (o ApplicationCommandInteractionDataOption) StringValue() string {
	if v, ok := o.Value.(string); ok {
		return v
	}
	return ""
}

// IntValue returns an integer option value.
func (o ApplicationCommandInteractionDataOption) IntValue() int64 {
	switch v := o.Value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

// BoolValue returns a boolean option value.
func (o ApplicationCommandInteractionDataOption) BoolValue() bool {
	v, _ := o.Value.(bool)
	return v
}

// FloatValue returns a numeric option value.
func (o ApplicationCommandInteractionDataOption) FloatValue() float64 {
	switch v := o.Value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case json.Number:
		n, _ := v.Float64()
		return n
	default:
		return 0
	}
}

// InteractionResolvedData contains entities resolved for a command interaction.
type InteractionResolvedData struct {
	Users       map[int64]User       `json:"users,omitempty"`
	Members     map[int64]Member     `json:"members,omitempty"`
	Roles       map[int64]Role       `json:"roles,omitempty"`
	Channels    map[int64]Channel    `json:"channels,omitempty"`
	Messages    map[int64]Message    `json:"messages,omitempty"`
	Attachments map[int64]Attachment `json:"attachments,omitempty"`
}

// Interaction is dispatched to bots through INTERACTION_CREATE.
type Interaction struct {
	ID                           int64                              `json:"id"`
	ApplicationID                int64                              `json:"application_id"`
	Type                         InteractionType                    `json:"type"`
	Data                         *ApplicationCommandInteractionData `json:"data,omitempty"`
	GuildID                      *int64                             `json:"guild_id,omitempty"`
	ChannelID                    *int64                             `json:"channel_id,omitempty"`
	Member                       map[string]any                     `json:"member,omitempty"`
	User                         *User                              `json:"user,omitempty"`
	AppPermissions               string                             `json:"app_permissions,omitempty"`
	Locale                       string                             `json:"locale,omitempty"`
	GuildLocale                  string                             `json:"guild_locale,omitempty"`
	Context                      *InteractionContextType            `json:"context,omitempty"`
	Token                        string                             `json:"token"`
	Version                      int                                `json:"version"`
	AuthorizingIntegrationOwners map[string]string                  `json:"authorizing_integration_owners,omitempty"`
}

// ApplicationCommandData returns interaction data or an empty value.
func (i *Interaction) ApplicationCommandData() ApplicationCommandInteractionData {
	if i == nil || i.Data == nil {
		return ApplicationCommandInteractionData{}
	}
	return *i.Data
}

// InteractionResponse is sent to acknowledge an interaction.
type InteractionResponse struct {
	Type InteractionResponseType  `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

// InteractionResponseData is the message, autocomplete, or modal response body.
type InteractionResponseData struct {
	Content     string                            `json:"content,omitempty"`
	Embeds      []Embed                           `json:"embeds,omitempty"`
	Flags       int                               `json:"flags,omitempty"`
	Choices     []*ApplicationCommandOptionChoice `json:"choices,omitempty"`
	CustomID    string                            `json:"custom_id,omitempty"`
	Title       string                            `json:"title,omitempty"`
	Components  []json.RawMessage                 `json:"components,omitempty"`
	Attachments []int64                           `json:"attachments,omitempty"`
}

// ApplicationCommands lists global application commands.
func (s *Session) ApplicationCommands(ctx context.Context, applicationID int64, options ...RequestOption) ([]*ApplicationCommand, error) {
	var out []*ApplicationCommand
	err := s.Request(ctx, http.MethodGet, EndpointApplicationCommands(formatID(applicationID)), nil, nil, &out, options...)
	return out, err
}

// GuildApplicationCommands lists guild-scoped application commands.
func (s *Session) GuildApplicationCommands(ctx context.Context, applicationID, guildID int64, options ...RequestOption) ([]*ApplicationCommand, error) {
	var out []*ApplicationCommand
	err := s.Request(ctx, http.MethodGet, EndpointGuildApplicationCommands(formatID(applicationID), formatID(guildID)), nil, nil, &out, options...)
	return out, err
}

// ApplicationCommand returns a command by ID. Pass guildID 0 for global commands.
func (s *Session) ApplicationCommand(ctx context.Context, applicationID, guildID, commandID int64, options ...RequestOption) (*ApplicationCommand, error) {
	path := EndpointApplicationCommand(formatID(applicationID), formatID(commandID))
	if guildID > 0 {
		path = EndpointGuildApplicationCommand(formatID(applicationID), formatID(guildID), formatID(commandID))
	}
	var out ApplicationCommand
	err := s.Request(ctx, http.MethodGet, path, nil, nil, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplicationCommandCreate creates a command. Pass guildID 0 for a global command.
func (s *Session) ApplicationCommandCreate(ctx context.Context, applicationID, guildID int64, command *ApplicationCommand, options ...RequestOption) (*ApplicationCommand, error) {
	path := EndpointApplicationCommands(formatID(applicationID))
	if guildID > 0 {
		path = EndpointGuildApplicationCommands(formatID(applicationID), formatID(guildID))
	}
	var out ApplicationCommand
	err := s.Request(ctx, http.MethodPost, path, nil, command, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplicationCommandEdit replaces a command. Pass guildID 0 for a global command.
func (s *Session) ApplicationCommandEdit(ctx context.Context, applicationID, guildID, commandID int64, command *ApplicationCommand, options ...RequestOption) (*ApplicationCommand, error) {
	path := EndpointApplicationCommand(formatID(applicationID), formatID(commandID))
	if guildID > 0 {
		path = EndpointGuildApplicationCommand(formatID(applicationID), formatID(guildID), formatID(commandID))
	}
	var out ApplicationCommand
	err := s.Request(ctx, http.MethodPatch, path, nil, command, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplicationCommandBulkOverwrite replaces all commands in a scope. Pass guildID 0 for global commands.
func (s *Session) ApplicationCommandBulkOverwrite(ctx context.Context, applicationID, guildID int64, commands []*ApplicationCommand, options ...RequestOption) ([]*ApplicationCommand, error) {
	path := EndpointApplicationCommands(formatID(applicationID))
	if guildID > 0 {
		path = EndpointGuildApplicationCommands(formatID(applicationID), formatID(guildID))
	}
	var out []*ApplicationCommand
	err := s.Request(ctx, http.MethodPut, path, nil, commands, &out, options...)
	return out, err
}

// ApplicationCommandDelete deletes a command. Pass guildID 0 for a global command.
func (s *Session) ApplicationCommandDelete(ctx context.Context, applicationID, guildID, commandID int64, options ...RequestOption) error {
	path := EndpointApplicationCommand(formatID(applicationID), formatID(commandID))
	if guildID > 0 {
		path = EndpointGuildApplicationCommand(formatID(applicationID), formatID(guildID), formatID(commandID))
	}
	return s.Request(ctx, http.MethodDelete, path, nil, nil, nil, options...)
}

// InteractionRespond acknowledges an interaction by using its ID and token.
func (s *Session) InteractionRespond(ctx context.Context, interaction *Interaction, response *InteractionResponse, options ...RequestOption) error {
	if interaction == nil {
		return ErrNilInteraction
	}
	return s.InteractionRespondByID(ctx, interaction.ID, interaction.Token, response, options...)
}

// InteractionRespondByID acknowledges an interaction by ID and token.
func (s *Session) InteractionRespondByID(ctx context.Context, interactionID int64, token string, response *InteractionResponse, options ...RequestOption) error {
	if response == nil {
		response = &InteractionResponse{}
	}
	return s.Request(ctx, http.MethodPost, EndpointInteractionCallback(interactionID, token), nil, response, nil, options...)
}

// InteractionResponse gets the original interaction response.
func (s *Session) InteractionResponse(ctx context.Context, applicationID int64, token string, options ...RequestOption) (*Message, error) {
	var out Message
	err := s.Request(ctx, http.MethodGet, EndpointOriginalInteractionResponse(applicationID, token), nil, nil, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// InteractionResponseEdit edits the original interaction response.
func (s *Session) InteractionResponseEdit(ctx context.Context, applicationID int64, token string, data *InteractionResponseData, options ...RequestOption) (*Message, error) {
	if data == nil {
		data = &InteractionResponseData{}
	}
	var out Message
	err := s.Request(ctx, http.MethodPatch, EndpointOriginalInteractionResponse(applicationID, token), nil, data, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// InteractionResponseDelete deletes the original interaction response.
func (s *Session) InteractionResponseDelete(ctx context.Context, applicationID int64, token string, options ...RequestOption) error {
	return s.Request(ctx, http.MethodDelete, EndpointOriginalInteractionResponse(applicationID, token), nil, nil, nil, options...)
}

// FollowupMessageCreate creates an interaction followup message.
func (s *Session) FollowupMessageCreate(ctx context.Context, applicationID int64, token string, data *InteractionResponseData, options ...RequestOption) (*Message, error) {
	if data == nil {
		data = &InteractionResponseData{}
	}
	var out Message
	err := s.Request(ctx, http.MethodPost, EndpointInteractionFollowup(applicationID, token), nil, data, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// FollowupMessageEdit edits an interaction followup message.
func (s *Session) FollowupMessageEdit(ctx context.Context, applicationID int64, token string, messageID int64, data *InteractionResponseData, options ...RequestOption) (*Message, error) {
	if data == nil {
		data = &InteractionResponseData{}
	}
	var out Message
	err := s.Request(ctx, http.MethodPatch, EndpointInteractionFollowupMessage(applicationID, token, messageID), nil, data, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// FollowupMessageDelete deletes an interaction followup message.
func (s *Session) FollowupMessageDelete(ctx context.Context, applicationID int64, token string, messageID int64, options ...RequestOption) error {
	return s.Request(ctx, http.MethodDelete, EndpointInteractionFollowupMessage(applicationID, token, messageID), nil, nil, nil, options...)
}
