package guild

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

const (
	ErrUnableToGetUserToken         = "unable to get user token"
	ErrUnableToGetGuildMemberToken  = "unable to get guild member token"
	ErrUnableToParseBody            = "unable to parse body"
	ErrPermissionsRequired          = "permissions required"
	ErrUnableToCreateAttachment     = "unable to create attachment"
	ErrUnableToCreateUploadURL      = "unable to create upload url"
	ErrIncorrectChannelID           = "incorrect channel ID"
	ErrIncorrectGuildID             = "incorrect guild ID"
	ErrIncorrectMemberID            = "incorrect member ID"
	ErrFileIsTooBig                 = "file is too big"
	ErrUnableToSendMessage          = "unable to send message"
	ErrUnableToGetUser              = "unable to get user"
	ErrUnableToGetUserDiscriminator = "unable to get discriminator"
	ErrUnableToGetAttachements      = "unable to get attachments"
	ErrUnableToCreateGuild          = "unable to create guild"
	ErrUnableToGetGuildMember       = "unable to get member"
	ErrUnableToGetGuildByID         = "unable to get guild by id"
	ErrUnableToUpdateGuild          = "unable to update guild"
	ErrUnableToGetRoles             = "unable to get roles"
	ErrUnableToCreateChannelGroup   = "unable to create channel group"
	ErrNotAMember                   = "not a member"

	// Validation error messages
	ErrGuildNameRequired   = "guild name is required"
	ErrGuildNameTooShort   = "guild name must be at least 2 characters"
	ErrGuildNameTooLong    = "guild name must be less than 100 characters"
	ErrChannelNameRequired = "channel name is required"
	ErrChannelNameTooShort = "channel name must be at least 2 characters"
	ErrChannelNameTooLong  = "channel name must be less than 100 characters"
	ErrChannelNameInvalid  = "channel name can only contain letters, numbers, hyphens, and underscores"
	ErrChannelTypeInvalid  = "invalid channel type"
	ErrIconIdInvalid       = "icon ID must be positive"
	ErrParentIdInvalid     = "parent ID must be positive"
	ErrPermissionsInvalid  = "permissions must be non-negative"
)

var (
	channelNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type CreateGuildRequest struct {
	Name   string `json:"name"`
	IconId *int64 `json:"icon_id"`
	Public bool   `json:"public"`
}

func (r CreateGuildRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrGuildNameRequired),
			validation.RuneLength(2, 0).Error(ErrGuildNameTooShort),
			validation.RuneLength(0, 100).Error(ErrGuildNameTooLong),
		),
		validation.Field(&r.IconId,
			validation.When(r.IconId != nil, validation.Min(int64(1)).Error(ErrIconIdInvalid)),
		),
	)
}

type UpdateGuildRequest struct {
	Name        *string `json:"name"`
	IconId      *int64  `json:"icon_id"`
	Public      *bool   `json:"public"`
	Permissions *int64  `json:"permissions"`
}

func (r UpdateGuildRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.When(r.Name != nil,
				validation.RuneLength(2, 0).Error(ErrGuildNameTooShort),
				validation.RuneLength(0, 100).Error(ErrGuildNameTooLong),
			),
		),
		validation.Field(&r.IconId,
			validation.When(r.IconId != nil, validation.Min(int64(1)).Error(ErrIconIdInvalid)),
		),
		validation.Field(&r.Permissions,
			validation.When(r.Permissions != nil, validation.Min(int64(0)).Error(ErrPermissionsInvalid)),
		),
	)
}

type CreateGuildChannelCategoryRequest struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func (r CreateGuildChannelCategoryRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrChannelNameRequired),
			validation.RuneLength(2, 0).Error(ErrChannelNameTooShort),
			validation.RuneLength(0, 100).Error(ErrChannelNameTooLong),
			validation.Match(channelNameRegex).Error(ErrChannelNameInvalid),
		),
	)
}

type CreateGuildChannelRequest struct {
	Name     string            `json:"name"`
	Type     model.ChannelType `json:"type"`
	ParentId *int64            `json:"parent_id"`
	Private  bool              `json:"private"`
}

func (r CreateGuildChannelRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrChannelNameRequired),
			validation.RuneLength(2, 0).Error(ErrChannelNameTooShort),
			validation.RuneLength(0, 100).Error(ErrChannelNameTooLong),
			validation.Match(channelNameRegex).Error(ErrChannelNameInvalid),
		),
		validation.Field(&r.Type,
			validation.Required,
			validation.In(
				model.ChannelTypeGuild,
				model.ChannelTypeGuildVoice,
				model.ChannelTypeGuildCategory,
				model.ChannelTypeDM,
				model.ChannelTypeGroupDM,
				model.ChannelTypeThread,
			).Error(ErrChannelTypeInvalid),
		),
		validation.Field(&r.ParentId,
			validation.When(r.ParentId != nil, validation.Min(int64(1)).Error(ErrParentIdInvalid)),
		),
	)
}

// Common data structures for guild operations
type guildContext struct {
	User   *helper.JWTUser
	Guild  *model.Guild
	Member *model.Member
}

type channelPermissionContext struct {
	User    *helper.JWTUser
	Guild   *model.Guild
	Channel *model.Channel
	Roles   map[int64]*model.Role
}

type memberRole struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Permissions int64  `json:"permissions"`
}

// DTO conversion functions

func channelModelToDTO(c *model.Channel, guildId *int64, position int) dto.Channel {
	return dto.Channel{
		Id:          c.Id,
		Type:        c.Type,
		GuildId:     guildId,
		Name:        c.Name,
		ParentId:    c.ParentID,
		Position:    position,
		Topic:       c.Topic,
		Permissions: c.Permissions,
		CreatedAt:   c.CreatedAt,
	}
}

// buildGuildDTO creates a guild DTO from model
func buildGuildDTO(guild *model.Guild) dto.Guild {
	return dto.Guild{
		Id:          guild.Id,
		Name:        guild.Name,
		Icon:        guild.Icon,
		Owner:       guild.OwnerId,
		Public:      guild.Public,
		Permissions: guild.Permissions,
	}
}

func roleModelToDTO(r model.Role) dto.Role {
	return dto.Role{
		Id:          r.Id,
		GuildId:     r.GuildId,
		Name:        r.Name,
		Color:       r.Color,
		Permissions: r.Permissions,
	}
}

func roleModelToDTOMany(roles []model.Role) []dto.Role {
	result := make([]dto.Role, len(roles))
	for i, r := range roles {
		result[i] = roleModelToDTO(r)
	}
	return result
}
