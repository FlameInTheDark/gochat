package user

import (
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

const (
	ErrUnableToGetUserToken        = "unable to get user token"
	ErrBadRequest                  = "incorrect request"
	ErrUnableToParseID             = "unable to parse id"
	ErrUnableToGetUser             = "unable to get user"
	ErrUnableToGetDiscriminator    = "unable to get user discriminator"
	ErrUnableToModifyUser          = "unable to modify user"
	ErrUnableToGetMember           = "unable to get member"
	ErrUnableToGetRoles            = "unable to get roles"
	ErrUnableToGetGuildByID        = "unable to get guild by id"
	ErrUnableToGetGuilds           = "unable to get guilds"
	ErrUnableToGetUserGuilds       = "unable to get user guilds"
	ErrUnableToLeaveOwnServer      = "unable to leave own guild"
	ErrUnableToRemoveMember        = "unable to remove member"
	ErrUnableToParseRequestBody    = "unable to parse request body"
	ErrUnableToCreateChannel       = "unable to create channel"
	ErrUnableToCreateDMChannel     = "unable to create dm channel"
	ErrUnableToGetChannel          = "unable to get channel"
	ErrUnableToGetDMChannel        = "unable to get dm channel"
	ErrUnableToGetGroupDMChannel   = "unable to get group dm channel"
	ErrUnableToJoingGroupDmChannel = "unable to join group dm channel"
)

type ModifyUserRequest struct {
	Avatar *int64  `json:"avatar,omitempty"`
	Name   *string `json:"Name,omitempty"`
}

type CreateDMRequest struct {
	RecipientId int64 `json:"recipient_id"`
}

type CreateDMManyRequest struct {
	ChannelId    *int64  `json:"channel_id"`
	RecipientsId []int64 `json:"recipients_id"`
}

func modelToUser(m model.User) dto.User {
	return dto.User{
		Id:     m.Id,
		Name:   m.Name,
		Avatar: m.Avatar,
	}
}

func guildModelToGuild(m model.Guild, user int64) dto.Guild {
	return dto.Guild{
		Id:     m.Id,
		Name:   m.Name,
		Icon:   m.Icon,
		Owner:  m.OwnerId == user,
		Public: m.Public,
	}
}

func guildModelToGuildMany(guilds []model.Guild, user int64) []dto.Guild {
	models := make([]dto.Guild, len(guilds))
	for i, g := range guilds {
		models[i] = guildModelToGuild(g, user)
	}
	return models
}
