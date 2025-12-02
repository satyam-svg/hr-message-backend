package models

import "time"

// Contact represents a contact extracted from PDF
type Contact struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	IsSent      bool      `json:"is_sent"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactResponse represents contact data in response
type ContactResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	IsSent      bool      `json:"is_sent"`
	CreatedAt   time.Time `json:"created_at"`
}

// SaveContactRequest represents the request to save a contact
type SaveContactRequest struct {
	Name        string `json:"name" validate:"required"`
	CompanyName string `json:"company_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
}
