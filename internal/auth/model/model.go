package model

import (
	"github.com/google/uuid"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type User struct {
	ID       uuid.UUID
	Username string
	Password string
	Role     RoleType
}

type RoleType string

const (
	RoleEditor RoleType = "editor"
	RoleViewer RoleType = "viewer"
)
