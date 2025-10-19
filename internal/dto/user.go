package dto

type User struct {
	Id            int64       `json:"id" example:"2230469276416868352"`
	Name          string      `json:"name" example:"FancyUserName"`
	Discriminator string      `json:"discriminator" example:"uniquename"`
	Avatar        *AvatarData `json:"avatar,omitempty"`
}
