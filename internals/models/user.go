package models

import "time"

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateEmailSettingsRequest represents the request payload for updating email settings
type UpdateEmailSettingsRequest struct {
	ProfessionalEmail string `json:"professional_email" validate:"required,email"`
	MailAppPassword   string `json:"mail_app_password" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse represents user data in response
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// ActivityResponse represents a single activity log
type ActivityResponse struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserProfileResponse represents full user profile data
type UserProfileResponse struct {
	ID                string             `json:"id"`
	Email             string             `json:"email"`
	Name              string             `json:"name"`
	ProfessionalEmail string             `json:"professional_email,omitempty"`
	MailAppPassword   string             `json:"mail_app_password,omitempty"`
	DailyLimit        int                `json:"daily_limit"`
	PdfUploadCount    int                `json:"pdf_upload_count"`
	EmailsSent        int                `json:"emails_sent"`
	CreatedAt         time.Time          `json:"created_at"`
	Contacts          []ContactResponse  `json:"contacts"`
	Activities        []ActivityResponse `json:"activities"`
	Template          *TemplateResponse  `json:"template,omitempty"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
