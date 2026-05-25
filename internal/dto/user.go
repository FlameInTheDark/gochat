package dto

type User struct {
	Id            int64       `json:"id" example:"2230469276416868352"`
	Name          string      `json:"name" example:"FancyUserName"`
	Discriminator string      `json:"discriminator" example:"uniquename"`
	Bio           *string     `json:"bio" example:"Building gochat one endpoint at a time"`
	BannerColor   *int        `json:"banner_color" example:"3447003"`
	PanelColor    *int        `json:"panel_color" example:"15158332"`
	Avatar        *AvatarData `json:"avatar,omitempty"`
	Banner        *BannerData `json:"banner,omitempty"`
	PersonalNote  *string     `json:"personal_note,omitempty" example:"Met during the release party"`
}
