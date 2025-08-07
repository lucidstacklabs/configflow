package admin

type InitRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PasswordChangeRequest struct {
	Password string `json:"password" binding:"required"`
}

type AdditionRequest struct {
	Username string `json:"username" binding:"required"`
}
