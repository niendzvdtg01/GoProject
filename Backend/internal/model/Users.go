package model

import "time"

const (
	RoleManager = "manager"
	RoleMember  = "member"
)

type User struct {
	UserID       string    `json:"userId" db:"user_id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type PublicUser struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		UserID:    u.UserID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func IsValidRole(role string) bool {
	return role == RoleManager || role == RoleMember
}
