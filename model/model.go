package model

import (
	"time"
)

type Plugin struct {
	Name       string `json:"name"`
	PluginName string `json:"plugin_name"`
	Type       string `json:"type"`
	IsMissing  bool   `json:"is_missing"`
}

type ResolutionResult struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Genre        string   `json:"genre"`
	Tags         []string `json:"tags"`
	Plugins      []Plugin `json:"plugins"`
	PluginsCount int      `json:"plugins_count"`
	Error        string   `json:"error,omitempty"`
}

type Project struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	AuthorID     string    `json:"author_id"`
	AuthorName   string    `json:"author"`
	UploadDate   time.Time `json:"upload_date"`
	Size         int64     `json:"size"`
	SizeDisplay  string    `json:"size_display"`
	PluginsCount int       `json:"plugins_count"`
	Description  string    `json:"description"`
	Genre        string    `json:"genre"`
	FilePath     string    `json:"-"`
	Plugins      []Plugin  `json:"plugins,omitempty" gorm:"serializer:json"`
	Tags         []Tag     `json:"tags,omitempty" gorm:"many2many:project_tags;"`
}

type Tag struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex"`
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

