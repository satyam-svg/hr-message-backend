package services

import (
	"fmt"

	"github.com/satyam-svg/hr_backend/internals/models"
	"gopkg.in/gomail.v2"
)

// EmailService handles email sending business logic
type EmailService struct{}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendEmail sends an email using Gmail SMTP
func (s *EmailService) SendEmail(req models.SendEmailRequest) error {
	// Create a new message
	m := gomail.NewMessage()

	// Set email headers
	m.SetHeader("From", req.SenderEmail)
	m.SetHeader("To", req.RecipientEmail)
	m.SetHeader("Subject", req.Subject)

	// Set email body
	m.SetBody("text/plain", req.Body)

	// Attach files if provided
	for _, filePath := range req.AttachmentPaths {
		m.Attach(filePath)
	}

	// Configure Gmail SMTP settings
	d := gomail.NewDialer("smtp.gmail.com", 587, req.SenderEmail, req.SenderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
