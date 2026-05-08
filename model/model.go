package model

type Plugin struct {
	Name       string `json:"name"`
	PluginName string `json:"plugin_name"`
	Type       string `json:"type"`
}

type ResolutionResult struct {
	Plugins []Plugin `json:"plugins"`
	Error   string   `json:"error,omitempty"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

