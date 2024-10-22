package user

import "github.com/FlameInTheDark/gochat/internal/database/model"

type UserResponse struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Determinator string `json:"determinator"`
	Avatar       *int64 `json:"avatar"`
}

func modelToUser(m model.User) UserResponse {
	return UserResponse{
		Id:           m.Id,
		Name:         m.Name,
		Determinator: m.Determinator,
		Avatar:       m.Avatar,
	}
}

type Guild struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Icon   *int64 `json:"icon"`
	Owner  bool   `json:"owner"`
	Public bool   `json:"public"`
}

func guildModelToGuild(m model.Guild, user int64) Guild {
	return Guild{
		Id:     m.Id,
		Name:   m.Name,
		Icon:   m.Icon,
		Owner:  m.OwnerId == user,
		Public: m.Public,
	}
}

func guildModelToGuildMany(guilds []model.Guild, user int64) []Guild {
	models := make([]Guild, len(guilds))
	for i, g := range guilds {
		models[i] = guildModelToGuild(g, user)
	}
	return models
}
