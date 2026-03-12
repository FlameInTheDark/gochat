package user

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/FlameInTheDark/gochat/internal/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/gofiber/fiber/v2"
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
	ErrUnableToGetReadStates         = "unable to get read states"
	ErrUnableToGetMembership         = "unable to get membership"

	// Validation error messages
	ErrUserNameTooShort           = "user name must be at least 4 characters"
	ErrUserNameTooLong            = "user name must be less than 20 characters"
	ErrAvatarIdInvalid            = "avatar ID must be positive"
	ErrRecipientIdRequired        = "recipient ID is required"
	ErrRecipientIdInvalid         = "recipient ID must be positive"
	ErrChannelIdInvalid           = "channel ID must be positive"
	ErrRecipientsRequired         = "at least one recipient is required"
	ErrRecipientsInvalid          = "recipient IDs must be positive"
	ErrTooManyRecipients          = "maximum 10 recipients allowed"
	ErrUnableToDeleteActiveAvatar = "unable to delete active avatar"
	ErrNoFieldsToUpdate           = "at least one field must be provided for update"
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
	Version             int64                            `json:"version"`
	Settings            *model.UserSettingsData          `json:"settings"`
	ContentHosts        []string                         `json:"content_hosts"`
	ReadStates          map[int64]int64                  `json:"read_states"`
	GuildsLastMessages  map[int64]map[int64]int64        `json:"guilds_last_messages"`
	ThreadsLastMessages map[int64]int64                  `json:"threads_last_messages"`
	JoinedThreads       map[int64]map[int64][]int64      `json:"joined_threads"` // Joined thread IDs grouped as guild_id -> parent_channel_id -> sorted thread ids.
	Guilds              []dto.Guild                      `json:"guilds"`
	GuildEmojis         map[int64][]dto.EmojiRef         `json:"guild_emojis"`
	Mentions            map[int64][]model.Mention        `json:"mentions,omitempty"`
	ChannelMentions     map[int64][]model.ChannelMention `json:"channel_mentions,omitempty"`
}

func modelToSettings(m *model.UserSettings, guilds []dto.Guild, guildEmojis map[int64][]dto.EmojiRef, rs map[int64]int64, glms map[int64]map[int64]int64) (UserSettingsResponse, error) {
	var settings model.UserSettingsData
	if len(m.Settings) > 0 {
		if err := json.Unmarshal(m.Settings, &settings); err != nil {
			return UserSettingsResponse{ReadStates: rs}, err
		}
	}
	return UserSettingsResponse{
		Version:             m.Version,
		Settings:            &settings,
		ContentHosts:        nil,
		ReadStates:          rs,
		GuildsLastMessages:  glms,
		ThreadsLastMessages: map[int64]int64{},
		JoinedThreads:       map[int64]map[int64][]int64{},
		Guilds:              guilds,
		GuildEmojis:         guildEmojis,
	}, nil
}

func filterGuildLastMessages(glms map[int64]map[int64]int64, channels []model.Channel) map[int64]map[int64]int64 {
	if len(glms) == 0 || len(channels) == 0 {
		return map[int64]map[int64]int64{}
	}

	allowedChannels := make(map[int64]struct{}, len(channels))
	for _, channel := range channels {
		if channel.Type == model.ChannelTypeThread {
			continue
		}
		allowedChannels[channel.Id] = struct{}{}
	}

	filtered := make(map[int64]map[int64]int64, len(glms))
	for guildID, channelMessages := range glms {
		for channelID, lastMessageID := range channelMessages {
			if _, ok := allowedChannels[channelID]; !ok {
				continue
			}
			if filtered[guildID] == nil {
				filtered[guildID] = make(map[int64]int64)
			}
			filtered[guildID][channelID] = lastMessageID
		}
	}

	return filtered
}

func filterThreadLastMessages(joined map[int64]struct{}, channels []model.Channel, glms map[int64]map[int64]int64) map[int64]int64 {
	if len(joined) == 0 || len(channels) == 0 || len(glms) == 0 {
		return map[int64]int64{}
	}

	liveThreads := make(map[int64]struct{}, len(joined))
	for _, channel := range channels {
		if channel.Type != model.ChannelTypeThread {
			continue
		}
		if _, ok := joined[channel.Id]; ok {
			liveThreads[channel.Id] = struct{}{}
		}
	}

	out := make(map[int64]int64, len(liveThreads))
	for _, channelMessages := range glms {
		for channelID, lastMessageID := range channelMessages {
			if _, ok := liveThreads[channelID]; !ok {
				continue
			}
			out[channelID] = lastMessageID
		}
	}
	return out
}

func buildJoinedThreads(joined map[int64]struct{}, channels []model.Channel, guildChannels []model.GuildChannel) map[int64]map[int64][]int64 {
	if len(joined) == 0 || len(channels) == 0 || len(guildChannels) == 0 {
		return map[int64]map[int64][]int64{}
	}

	guildByChannel := make(map[int64]int64, len(guildChannels))
	for _, guildChannel := range guildChannels {
		guildByChannel[guildChannel.ChannelId] = guildChannel.GuildId
	}

	out := make(map[int64]map[int64][]int64)
	for _, channel := range channels {
		if channel.Type != model.ChannelTypeThread || channel.ParentID == nil {
			continue
		}
		if _, ok := joined[channel.Id]; !ok {
			continue
		}
		guildID, ok := guildByChannel[channel.Id]
		if !ok {
			continue
		}
		if out[guildID] == nil {
			out[guildID] = make(map[int64][]int64)
		}
		parentID := *channel.ParentID
		out[guildID][parentID] = append(out[guildID][parentID], channel.Id)
	}

	for guildID, channelsMap := range out {
		for parentID, threads := range channelsMap {
			sort.Slice(threads, func(i, j int) bool {
				return threads[i] < threads[j]
			})
			channelsMap[parentID] = threads
		}
		out[guildID] = channelsMap
	}

	return out
}

func modelToUser(m model.User) dto.User {
	return dto.User{
		Id:   m.Id,
		Name: m.Name,
	}
}

func (e *entity) guildModelToGuild(c *fiber.Ctx, m model.Guild) dto.Guild {
	g := dto.Guild{
		Id:     m.Id,
		Name:   m.Name,
		Owner:  m.OwnerId,
		Public: m.Public,
	}
	if m.Icon != nil {
		key := fmt.Sprintf("icons:%d:%d", m.Id, *m.Icon)
		var cached dto.Icon
		if err := e.cache.GetJSON(c.UserContext(), key, &cached); err == nil && cached.URL != "" {
			g.Icon = &cached
			return g
		}

		if ic, err := e.icon.GetIcon(c.UserContext(), *m.Icon, m.Id); err == nil && ic.URL != nil {
			var w, h, size int64
			if ic.Width != nil {
				w = *ic.Width
			}
			if ic.Height != nil {
				h = *ic.Height
			}
			size = ic.FileSize
			var urlStr string
			if ic.URL != nil {
				urlStr = *ic.URL
			}
			ico := dto.Icon{Id: *m.Icon, URL: urlStr, Filesize: size, Width: w, Height: h}
			g.Icon = &ico
			_ = e.cache.SetJSON(c.UserContext(), key, ico)
		}
	}
	return g
}

func (e *entity) guildModelToGuildMany(c *fiber.Ctx, guilds []model.Guild) []dto.Guild {
	models := make([]dto.Guild, len(guilds))
	for i, g := range guilds {
		models[i] = e.guildModelToGuild(c, g)
	}
	return models
}

// Friend-related errors
const (
	ErrUnableToGetFriends           = "unable to get friends"
	ErrUnableToCreateFriendRequest  = "unable to create friend request"
	ErrUnableToRemoveFriend         = "unable to remove friend"
	ErrUnableToGetFriendRequests    = "unable to get friend requests"
	ErrUnableToAcceptFriendRequest  = "unable to accept friend request"
	ErrUnableToDeclineFriendRequest = "unable to decline friend request"
)

// Payloads for friend operations
type CreateFriendRequestRequest struct {
	Discriminator string `json:"discriminator"`
}

func (r CreateFriendRequestRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Discriminator,
			validation.Required.Error("discriminator is required"),
		),
	)
}

type UnfriendRequest struct {
	UserId int64 `json:"user_id,string"`
}

func (r UnfriendRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserId,
			validation.Required.Error("user_id is required"),
			validation.Min(int64(1)).Error("user_id must be positive"),
		),
	)
}

type FriendRequestAction struct {
	UserId int64 `json:"user_id,string"`
}

func (r FriendRequestAction) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserId,
			validation.Required.Error("user_id is required"),
			validation.Min(int64(1)).Error("user_id must be positive"),
		),
	)
}

// usersWithDiscriminators converts users list and discriminator list into DTO users
func usersWithDiscriminators(users []model.User, discs []model.Discriminator) []dto.User {
	discMap := make(map[int64]string, len(discs))
	for _, d := range discs {
		discMap[d.UserId] = d.Discriminator
	}
	res := make([]dto.User, len(users))
	for i, u := range users {
		res[i] = dto.User{
			Id:            u.Id,
			Name:          u.Name,
			Discriminator: discMap[u.Id],
		}
	}
	return res
}

// DM channels last messages request
type DMChannelsLastRequest struct {
	ChannelIds []helper.StringInt64Array `json:"channel_ids"`
}

func (r DMChannelsLastRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelIds,
			validation.Required.Error(ErrRecipientsRequired),
			validation.Length(1, 0).Error(ErrRecipientsRequired),
			validation.Each(validation.Min(int64(1)).Error("channel_id must be positive")),
		),
	)
}

func dmChannelModelToDTO(cn *model.Channel, last map[int64]int64, participant *int64) dto.Channel {
	lm := cn.LastMessage
	if v, ok := last[cn.Id]; ok {
		lm = v
	}
	return dto.Channel{
		Id:            cn.Id,
		Type:          cn.Type,
		GuildId:       nil,
		ParticipantId: participant,
		Name:          cn.Name,
		ParentId:      cn.ParentID,
		Position:      0,
		Topic:         cn.Topic,
		Permissions:   cn.Permissions,
		Private:       cn.Private,
		Roles:         nil,
		LastMessageId: lm,
		CreatedAt:     cn.CreatedAt,
	}
}
