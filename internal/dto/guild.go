package dto

type Guild struct {
	Id          int64  `json:"id" example:"2230469276416868352"`    // Guild ID
	Name        string `json:"name" example:"My Guild"`             // Guild Name
	Icon        *int64 `json:"icon" example:"2230469276416868352"`  // Icon ID
	Owner       int64  `json:"owner" example:"2230469276416868352"` // Owner ID
	Public      bool   `json:"public" default:"false"`              // Whether the guild is public
	Permissions int64  `json:"permissions" default:"7927905"`       // Default guild Permissions. Check the permissions documentation for more info.
}
