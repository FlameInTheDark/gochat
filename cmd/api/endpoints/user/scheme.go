package user

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

const (
	ErrUnableToGetUserToken          = "unable to get user token"
	ErrBadRequest                    = "incorrect request"
	ErrUnableToParseID               = "unable to parse id"
	ErrUnableToGetUser               = "unable to get user"
	ErrUnableToGetDiscriminator      = "unable to get user discriminator"
	ErrUnableToModifyUser            = "unable to modify user"
	ErrUnableToGetMember             = "unable to get member"
	ErrUnableToGetRoles              = "unable to get roles"
	ErrUnableToGetGuildByID          = "unable to get guild by id"
	ErrUnableToGetGuilds             = "unable to get guilds"
	ErrUnableToGetUserGuilds         = "unable to get user guilds"
	ErrUnableToLeaveOwnServer        = "unable to leave own guild"
	ErrUnableToRemoveMember          = "unable to remove member"
	ErrUnableToParseRequestBody      = "unable to parse request body"
	ErrUnableToCreateChannel         = "unable to create channel"
	ErrUnableToCreateDMChannel       = "unable to create dm channel"
	ErrUnableToGetChannel            = "unable to get channel"
	ErrUnableToGetDMChannel          = "unable to get dm channel"
	ErrUnableToGetGroupDMChannel     = "unable to get group dm channel"
	ErrUnableToJoingGroupDmChannel   = "unable to join group dm channel"
	ErrUnableToGetUserSettings       = "unable to get user settings"
	ErrUnableToSetUserSettings       = "unable to set user settings"
	ErrUnableToUnmarshalUserSettings = "unable to unmarshal user settings"
	ErrUnableToParseVersion          = "unable to parse version"

	// Validation error messages
	ErrUserNameTooShort    = "user name must be at least 4 characters"
	ErrUserNameTooLong     = "user name must be less than 20 characters"
	ErrAvatarIdInvalid     = "avatar ID must be positive"
	ErrRecipientIdRequired = "recipient ID is required"
	ErrRecipientIdInvalid  = "recipient ID must be positive"
	ErrChannelIdInvalid    = "channel ID must be positive"
	ErrRecipientsRequired  = "at least one recipient is required"
	ErrRecipientsInvalid   = "recipient IDs must be positive"
	ErrTooManyRecipients   = "maximum 10 recipients allowed"
	ErrNoFieldsToUpdate    = "at least one field must be provided for update"
)

type ModifyUserRequest struct {
	Avatar *int64  `json:"avatar,omitempty" example:"2230469276416868352"` // Avatar ID.
	Name   *string `json:"name,omitempty" example:"NewFancyName"`          // User name.
}

func (r ModifyUserRequest) Validate() error {
	// Check if at least one field is provided
	if r.Avatar == nil && r.Name == nil {
		return validation.NewError("VALIDATION_NO_FIELDS", ErrNoFieldsToUpdate)
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.When(r.Name != nil,
				validation.RuneLength(4, 0).Error(ErrUserNameTooShort),
				validation.RuneLength(0, 20).Error(ErrUserNameTooLong),
			),
		),
		validation.Field(&r.Avatar,
			validation.When(r.Avatar != nil, validation.Min(int64(1)).Error(ErrAvatarIdInvalid)),
		),
	)
}

type CreateDMRequest struct {
	RecipientId int64 `json:"recipient_id"`
}

func (r CreateDMRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RecipientId,
			validation.Required.Error(ErrRecipientIdRequired),
			validation.Min(int64(1)).Error(ErrRecipientIdInvalid),
		),
	)
}

type CreateDMManyRequest struct {
	ChannelId    *int64  `json:"channel_id"`
	RecipientsId []int64 `json:"recipients_id"`
}

func (r CreateDMManyRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelId,
			validation.When(r.ChannelId != nil, validation.Min(int64(1)).Error(ErrChannelIdInvalid)),
		),
		validation.Field(&r.RecipientsId,
			validation.Required.Error(ErrRecipientsRequired),
			validation.Length(1, 0).Error(ErrRecipientsRequired),
			validation.Length(0, 10).Error(ErrTooManyRecipients),
			validation.Each(validation.Min(int64(1)).Error(ErrRecipientsInvalid)),
		),
	)
}

type UserSettingsResponse struct {
	Version  int64                   `json:"version"`
	Settings *model.UserSettingsData `json:"settings"`
}

func modelToSettings(m *model.UserSettings) (UserSettingsResponse, error) {
	var settings model.UserSettingsData
	err := json.Unmarshal(m.Settings, &settings)
	if err != nil {
		return UserSettingsResponse{}, err
	}
	return UserSettingsResponse{
		Version:  m.Version,
		Settings: &settings,
	}, nil
}

func modelToUser(m model.User) dto.User {
	return dto.User{
		Id:     m.Id,
		Name:   m.Name,
		Avatar: m.Avatar,
	}
}

func guildModelToGuild(m model.Guild) dto.Guild {
	return dto.Guild{
		Id:     m.Id,
		Name:   m.Name,
		Icon:   m.Icon,
		Owner:  m.OwnerId,
		Public: m.Public,
	}
}

func guildModelToGuildMany(guilds []model.Guild) []dto.Guild {
	models := make([]dto.Guild, len(guilds))
	for i, g := range guilds {
		models[i] = guildModelToGuild(g)
	}
	return models
}
