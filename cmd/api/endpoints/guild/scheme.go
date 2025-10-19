package guild

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
)

const (
	ErrUnableToGetUserToken            = "unable to get user token"
	ErrUnableToGetGuildMemberToken     = "unable to get guild member token"
	ErrUnableToGetGuildMembers         = "unable to get guild members"
	ErrUnableToGetUsersRoles           = "unable to get users roles"
	ErrUnableToGetGuildMembersProfiles = "unable to get guild members profiles"
	ErrUnableToParseBody               = "unable to parse body"
	ErrPermissionsRequired             = "permissions required"
	ErrUnableToCreateAttachment        = "unable to create attachment"
	ErrUnableToCreateUploadURL         = "unable to create upload url"
	ErrIncorrectChannelID              = "incorrect channel ID"
	ErrIncorrectGuildID                = "incorrect guild ID"
	ErrIncorrectMemberID               = "incorrect member ID"
	ErrIncorrectInviteID               = "incorrect invite ID"
	ErrIncorrectRoleID                 = "incorrect role ID"
	ErrIncorrectIconID                 = "incorrect icon ID"
	ErrFileIsTooBig                    = "file is too big"
	ErrUnableToSendMessage             = "unable to send message"
	ErrUnableToGetUser                 = "unable to get user"
	ErrUnableToGetUsers                = "unable to get users"
	ErrUnableToGetUserDiscriminator    = "unable to get discriminator"
	ErrUnableToGetDiscriminators       = "unable to get discriminators"
	ErrUnableToGetAttachements         = "unable to get attachments"
	ErrUnableToCreateGuild             = "unable to create guild"
	ErrUnableToGetGuildMember          = "unable to get member"
	ErrUnableToGetDiscriminator        = "unable to get discriminator"
	ErrUnableToGetGuildByID            = "unable to get guild by id"
	ErrUnableToUpdateGuild             = "unable to update guild"
	ErrUnableToDeleteGuild             = "unable to delete guild"
	ErrUnableToGetRoles                = "unable to get roles"
	ErrUnableToSetUserRole             = "unable to set user role"
	ErrUnableToRemoveUserRole          = "unable to remove user role"
	ErrUnableToCreateChannelGroup      = "unable to create channel group"
	ErrUnableToGetChannel              = "unable to get channel"
	ErrUnableToUpdateChannel           = "unable to update channel"
	ErrUnableToSetParentAsSelf         = "unable to set parent as self"
	ErrUnableToSetParentForCategory    = "unable to set parent for category"
	ErrNotAMember                      = "not a member"
	ErrUnableToGetReadState            = "unable to get read state"
	ErrUnableToSetReadState            = "unable to set read state"
	// Channel role permissions
	ErrUnableToGetChannelRolePerms = "unable to get channel role permissions"
	ErrUnableToSetChannelRolePerm  = "unable to set channel role permission"
	ErrUnableToUpdateChannelRole   = "unable to update channel role permission"
	ErrUnableToRemoveChannelRole   = "unable to remove channel role permission"
	// Invites
	ErrUnableToCreateInvite = "unable to create invite"
	ErrUnableToGetInvites   = "unable to get invites"
	ErrUnableToDeleteInvite = "unable to delete invite"
	ErrInviteNotFound       = "invite not found"
	ErrInviteCodeInvalid    = "invalid invite code"
	ErrRoleNotInGuild       = "role does not belong to this guild"

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
	// Roles
	ErrRoleNameRequired         = "role name is required"
	ErrRoleNameTooShort         = "role name must be at least 2 characters"
	ErrRoleNameTooLong          = "role name must be less than 100 characters"
	ErrRoleColorInvalid         = "role color must be between 0 and 16777215"
	ErrUnableToDeleteActiveIcon = "unable to delete active icon"
)

var (
	channelNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type CreateGuildRequest struct {
	Name   string `json:"name" example:"My unique guild"`        // Guild name
	IconId *int64 `json:"icon_id" example:"2230469276416868352"` // Icon ID
	Public bool   `json:"public" default:"false"`                // Whether the guild is public
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
	Name        *string `json:"name" example:"New guild name"`         // Guild name
	IconId      *int64  `json:"icon_id" example:"2230469276416868352"` // Icon ID
	Public      *bool   `json:"public" default:"false"`                // Whether the guild is public
	Permissions *int64  `json:"permissions" default:"7927905"`         // Permissions. Check the permissions documentation for more info.
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
	Name     string `json:"name" example:"category-name"` // Category channel name
	Private  bool   `json:"private" default:"false"`      // Whether the category channel is private. Private channels can only be seen by users with roles assigned to this channel.
	Position int    `json:"position" default:"0"`         // Channel position in the list. Should be set as the last position in the channel list, or it will be one of the first in the list.
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
	Name     string            `json:"name" example:"channel-name"`             // Channel name
	Type     model.ChannelType `json:"type" example:"0"`                        // Channel type
	ParentId *int64            `json:"parent_id" example:"2230469276416868352"` // Parent channel ID. A Parent channel can only be a category channel.
	Private  bool              `json:"private" default:"false"`                 // Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel.
	Position int               `json:"position" default:"0"`                    // Channel position in the list. Should be set as the last position in the channel list, or it will be one of the first in the list.
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

type ChannelOrder struct {
	Id       int64 `json:"id" example:"2230469276416868352"` // Channel ID.
	Position int   `json:"position" example:"4"`             // New channel position.
}

type PatchGuildChannelOrderRequest struct {
	Channels []ChannelOrder `json:"channels"` // List of channels to change order.
}

func (c ChannelOrder) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Id, validation.Required),
	)
}

func (r PatchGuildChannelOrderRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Channels,
			validation.Required,
			validation.Each(validation.By(func(v interface{}) error {
				co, ok := v.(ChannelOrder)
				if !ok {
					return validation.NewError("validation", "invalid channel element")
				}
				return co.Validate()
			})),
		),
	)
}

type PatchGuildChannelRequest struct {
	Name     *string `json:"name,omitempty" example:"new-channel-name"`         // Channel name.
	ParentId *int64  `json:"parent_id,omitempty" example:"2230469276416868352"` // Parent channel ID. A Parent channel can only be a category channel.
	Private  *bool   `json:"private,omitempty" default:"false"`                 // Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel.
	Topic    *string `json:"topic,omitempty" example:"Just a channel topic"`    // Channel topic.
}

func (r PatchGuildChannelRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.When(r.Name != nil,
				validation.RuneLength(2, 0).Error(ErrChannelNameTooShort),
				validation.RuneLength(0, 100).Error(ErrChannelNameTooLong),
				validation.Match(channelNameRegex).Error(ErrChannelNameInvalid),
			),
		),
		validation.Field(&r.ParentId, validation.When(r.ParentId != nil,
			validation.Min(int64(1)).Error(ErrParentIdInvalid)),
		),
	)
}

// Invites
type CreateInviteRequest struct {
	ExpiresInSec *int `json:"expires_in_sec" example:"86400"` // Expiration time in seconds. 0 means unlimited.
}

func (r CreateInviteRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ExpiresInSec,
			validation.When(r.ExpiresInSec != nil,
				validation.By(func(v interface{}) error {
					p, _ := v.(*int)
					if p == nil {
						return nil
					}
					if *p == 0 {
						return nil // unlimited is allowed
					}
					if *p < 60 {
						return validation.NewError("validation", "expires_in_sec must be 0 or >= 60 seconds")
					}
					if *p > 60*60*24*30 {
						return validation.NewError("validation", "expires_in_sec must be 0 or <= 2592000 seconds")
					}
					return nil
				}),
			),
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
func channelModelToDTO(c *model.Channel, guildId *int64, position int, roles []int64) dto.Channel {
	return dto.Channel{
		Id:            c.Id,
		Type:          c.Type,
		GuildId:       guildId,
		Name:          c.Name,
		ParentId:      c.ParentID,
		Position:      position,
		Topic:         c.Topic,
		Permissions:   c.Permissions,
		CreatedAt:     c.CreatedAt,
		Private:       c.Private,
		Roles:         roles,
		LastMessageId: c.LastMessage,
	}
}

// buildGuildDTO creates a guild DTO from model
func buildGuildDTO(guild *model.Guild) dto.Guild {
	return dto.Guild{
		Id:          guild.Id,
		Name:        guild.Name,
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

// Channel role permissions
type ChannelRolePermission struct {
	RoleId int64 `json:"role_id" example:"2230469276416868352"` // Role ID
	Accept int64 `json:"accept" example:"0"`                    // Allowed permission bits mask
	Deny   int64 `json:"deny" example:"0"`                      // Denied permission bits mask
}

type ChannelRolePermissionRequest struct {
	Accept int64 `json:"accept" example:"0"` // Allowed permission bits mask
	Deny   int64 `json:"deny" example:"0"`   // Denied permission bits mask
}

func (r ChannelRolePermissionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Accept, validation.Min(int64(0)).Error(ErrPermissionsInvalid)),
		validation.Field(&r.Deny, validation.Min(int64(0)).Error(ErrPermissionsInvalid)),
	)
}

type CreateGuildRoleRequest struct {
	Name        string `json:"name" example:"New Role"`  // Role name
	Color       int    `json:"color" example:"16777215"` // RGB int value
	Permissions int64  `json:"permissions" default:"0"`  // Permissions bitset
}

func (r CreateGuildRoleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrRoleNameRequired),
			validation.RuneLength(2, 0).Error(ErrRoleNameTooShort),
			validation.RuneLength(0, 100).Error(ErrRoleNameTooLong),
		),
		validation.Field(&r.Color,
			validation.Min(0).Error(ErrRoleColorInvalid),
			validation.Max(16777215).Error(ErrRoleColorInvalid),
		),
		validation.Field(&r.Permissions,
			validation.Min(int64(0)).Error(ErrPermissionsInvalid),
		),
	)
}

type PatchGuildRoleRequest struct {
	Name        *string `json:"name,omitempty" example:"Moderators"` // Role name
	Color       *int    `json:"color,omitempty" example:"16711680"`  // RGB int value
	Permissions *int64  `json:"permissions,omitempty"`               // Permissions bitset
}

func (r PatchGuildRoleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.When(r.Name != nil,
				validation.RuneLength(2, 0).Error(ErrRoleNameTooShort),
				validation.RuneLength(0, 100).Error(ErrRoleNameTooLong),
			),
		),
		validation.Field(&r.Color,
			validation.When(r.Color != nil,
				validation.Min(0).Error(ErrRoleColorInvalid),
				validation.Max(16777215).Error(ErrRoleColorInvalid),
			),
		),
		validation.Field(&r.Permissions,
			validation.When(r.Permissions != nil, validation.Min(int64(0)).Error(ErrPermissionsInvalid)),
		),
	)
}

func userToDTO(user model.User, dsc string) dto.User {
	return dto.User{
		Id:            user.Id,
		Name:          user.Name,
		Discriminator: dsc,
	}
}

func membersToDTO(members []model.Member, users []model.User, roles []model.UserRoles, dscs []model.Discriminator, avData map[int64]*dto.AvatarData) []dto.Member {
	var data = make([]dto.Member, len(members))
	for i, m := range members {
		u := userToDTO(users[i], dscs[i].Discriminator)
		if ad, ok := avData[m.UserId]; ok {
			u.Avatar = ad
		}
		data[i] = dto.Member{
			User:     u,
			Username: m.Username,
			Avatar:   m.Avatar,
			JoinAt:   m.JoinAt,
			Roles:    roles[i].Roles,
		}
	}
	return data
}

// CreateIconRequest is a request to create guild icon metadata
type CreateIconRequest struct {
	FileSize    int64  `json:"file_size" example:"120000"`
	ContentType string `json:"content_type" example:"image/png"`
}

func (r CreateIconRequest) Validate() error {
	if r.FileSize <= 0 || r.FileSize > 250*1024 {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}
	if len(r.ContentType) < 6 || r.ContentType[:6] != "image/" {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content type")
	}
	return nil
}
