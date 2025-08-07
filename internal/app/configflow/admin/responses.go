package admin

type TokenResponse struct {
	Token string `json:"token"`
}

type PasswordResponse struct {
	Admin    *Admin `json:"admin"`
	Password string `json:"password"`
}
