package dto

type User struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
	Avatar        *int64 `json:"avatar,omitempty"`
}
