package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // 'admin' or 'mahasiswa'
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID             int             `json:"id"`
	Email          string          `json:"email"`
	Role           string          `json:"role"`
	StudentID      *int            `json:"student_id,omitempty"`
	StudentProfile *StudentDetail  `json:"student_profile,omitempty"`
}

type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID *int   `json:"student_id,omitempty"`
}
