package domain

type User struct {
	DisplayName string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
}
