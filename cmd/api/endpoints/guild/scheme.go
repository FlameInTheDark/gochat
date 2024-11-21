package guild

import (
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

const (
	ErrUnableToGetUserToken         = "unable to get user token"
	ErrUnableToParseBody            = "unable to parse body"
	ErrPermissionsRequired          = "permissions required"
	ErrUnableToCreateAttachment     = "unable to create attachment"
	ErrUnableToCreateUploadURL      = "unable to create upload url"
	ErrIncorrectChannelID           = "incorrect channel ID"
	ErrIncorrectGuildID             = "incorrect guild ID"
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
)

type CreateGuildRequest struct {
	Name   string `json:"name"`
	IconId *int64 `json:"icon_id"`
	Public bool   `json:"public"`
}

type UpdateGuildRequest struct {
	Name   *string `json:"name"`
	IconId *int64  `json:"icon_id"`
	Public *bool   `json:"public"`
}

type CreateGuildChannelCategoryRequest struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type CreateGuildChannelRequest struct {
	Name     string            `json:"name"`
	Type     model.ChannelType `json:"type"`
	ParentId *int64            `json:"parent_id"`
	Private  bool              `json:"private"`
}

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
