package types

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID  int64  `json:"user_id"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}
