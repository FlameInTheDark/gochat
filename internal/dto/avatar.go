package dto

type AvatarUpload struct {
	Id     int64 `json:"id" example:"2230469276416868352"`
	UserId int64 `json:"user_id" example:"2230469276416868352"`
}

// AvatarData is returned with user profile to describe active avatar
type AvatarData struct {
	Id          int64   `json:"id" example:"2230469276416868352"`
	URL         string  `json:"url" example:"https://cdn.example.com/avatars/2230/2231.webp"`
	ContentType *string `json:"content_type,omitempty" example:"image/webp"`
	Width       *int64  `json:"width,omitempty" example:"128"`
	Height      *int64  `json:"height,omitempty" example:"128"`
	Size        int64   `json:"size" example:"245678"`
}

// Avatar represents a stored avatar item with its ID and metadata
type Avatar struct {
	Id          int64   `json:"id" example:"2230469276416868352"`
	URL         string  `json:"url" example:"https://cdn.example.com/avatars/2230/2231.webp"`
	ContentType *string `json:"content_type,omitempty" example:"image/webp"`
	Width       *int64  `json:"width,omitempty" example:"128"`
	Height      *int64  `json:"height,omitempty" example:"128"`
	Size        int64   `json:"size" example:"245678"`
}

type BannerUpload struct {
	Id     int64 `json:"id" example:"2230469276416868352"`
	UserId int64 `json:"user_id" example:"2230469276416868352"`
}

// BannerData is returned with user profiles to describe the active profile banner.
type BannerData struct {
	Exists      bool    `json:"exists" example:"true"`
	Id          int64   `json:"id,omitempty" example:"2230469276416868352"`
	URL         string  `json:"url,omitempty" example:"https://cdn.example.com/banners/2230/2231.webp"`
	ContentType *string `json:"content_type,omitempty" example:"image/webp"`
	Width       *int64  `json:"width,omitempty" example:"680"`
	Height      *int64  `json:"height,omitempty" example:"240"`
	Size        int64   `json:"size,omitempty" example:"245678"`
}
