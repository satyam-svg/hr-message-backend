package models

// SendEmailRequest represents the request payload for sending an email
type SendEmailRequest struct {
	SenderEmail     string   `json:"sender_email" validate:"required,email"`
	SenderPassword  string   `json:"sender_password" validate:"required"`
	RecipientEmail  string   `json:"recipient_email"`
	Subject         string   `json:"subject" validate:"required"`
	Body            string   `json:"body" validate:"required"`
	SendToAll       bool     `json:"send_to_all"`                // If true, sends to all contacts
	AttachmentPaths []string `json:"attachment_paths,omitempty"` // Optional file paths for attachments
}

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
