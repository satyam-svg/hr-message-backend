package models

import "time"

// Template represents an email template
type Template struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateTemplateRequest represents the request to update a template
type UpdateTemplateRequest struct {
	Name    string `json:"name" validate:"required"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

// TemplateResponse represents the template data in response
type TemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnhanceTemplateRequest represents the request to enhance a template
type EnhanceTemplateRequest struct {
	Content string `json:"content" validate:"required"`
}
