package auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email string `json:"email"`
}

type ConfirmationRequest struct {
	Id           int64  `json:"id"`
	Token        string `json:"token"`
	Name         string `json:"name"`
	Determinator string `json:"determinator"`
	Password     string `json:"password"`
}
