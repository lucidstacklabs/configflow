package apikey

type CreationRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateRequest struct {
	Name string `json:"name"`
}
