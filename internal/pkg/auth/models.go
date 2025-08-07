package auth

type AuthenticatedAdmin struct {
	ID string `json:"id"`
}

type AuthenticatedApiKey struct {
	ID string `json:"id"`
}
