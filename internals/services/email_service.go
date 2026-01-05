package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/prisma/db"
	"gopkg.in/gomail.v2"
)

// EmailService handles email sending business logic
type EmailService struct {
	client *db.PrismaClient
}

// NewEmailService creates a new email service
func NewEmailService(client *db.PrismaClient) *EmailService {
	return &EmailService{
		client: client,
	}
}

// StartEmailCampaign starts a background process to send emails to unsent contacts
func (s *EmailService) StartEmailCampaign(userId string) error {
	ctx := context.Background()

	// 1. Fetch User (for credentials)
	user, err := s.client.User.FindUnique(
		db.User.ID.Equals(userId),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// Check if user has email credentials
	professionalEmail, ok1 := user.ProfessionalEmail()
	mailAppPassword, ok2 := user.MailAppPassword()
	if !ok1 || !ok2 || professionalEmail == "" || mailAppPassword == "" {
		return fmt.Errorf("user email credentials not configured")
	}

	// 2. Fetch Template
	template, err := s.client.Template.FindUnique(
		db.Template.UserID.Equals(userId),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch template: %w", err)
	}

	// 3. Fetch Unsent Contacts
	contacts, err := s.client.Contact.FindMany(
		db.Contact.UserID.Equals(userId),
		db.Contact.IsSent.Equals(false),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch contacts: %w", err)
	}

	if len(contacts) == 0 {
		return nil // Nothing to send
	}

	// 4. Start Background Process
	go func() {
		// Create a new context for the background job
		bgCtx := context.Background()

		for _, contact := range contacts {
			// Double check if contact is still unsent (in case of race conditions or manual updates)
			currentContact, err := s.client.Contact.FindUnique(
				db.Contact.ID.Equals(contact.ID),
			).Exec(bgCtx)
			if err != nil || currentContact.IsSent {
				continue
			}

			// Prepare Email Body
			body := template.Body
			body = strings.ReplaceAll(body, "{name}", contact.Name)
			body = strings.ReplaceAll(body, "{company}", contact.CompanyName)
			body = strings.ReplaceAll(body, "{hr_name}", user.Name)

			// Prepare Request
			req := models.SendEmailRequest{
				SenderEmail:    professionalEmail,
				SenderPassword: mailAppPassword,
				RecipientEmail: contact.Email,
				Subject:        template.Subject,
				Body:           body,
			}

			// Send Email
			err = s.SendEmail(req)
			if err != nil {
				fmt.Printf("Failed to send email to %s: %v\n", contact.Email, err)
				// Continue to next contact even if one fails
			} else {
				// Update Contact Status
				_, err := s.client.Contact.FindUnique(
					db.Contact.ID.Equals(contact.ID),
				).Update(
					db.Contact.IsSent.Set(true),
				).Exec(bgCtx)
				if err != nil {
					fmt.Printf("Failed to update status for %s: %v\n", contact.Email, err)
				}
			}

			// Wait for 2 minutes before next email
			time.Sleep(2 * time.Minute)
		}
	}()

	return nil
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
	// Convert newlines to HTML breaks for proper rendering
	htmlBody := strings.ReplaceAll(req.Body, "\n", "<br>")
	m.SetBody("text/html", htmlBody)

	// Attach files if provided
	for _, filePath := range req.AttachmentPaths {
		m.Attach(filePath)
	}

	// Configure Gmail SMTP settings
	// Use port 465 for SSL (which works better on Render/Cloud) instead of 587 (TLS)
	d := gomail.NewDialer("smtp.gmail.com", 465, req.SenderEmail, req.SenderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmailBackground sends an email in the background
func (s *EmailService) SendEmailBackground(userId string, req models.SendEmailRequest) error {
	go func() {
		if req.SendToAll {
			// Fetch all contacts for the user
			ctx := context.Background()
			contacts, err := s.client.Contact.FindMany(
				db.Contact.UserID.Equals(userId),
			).Exec(ctx)
			if err != nil {
				fmt.Printf("Error fetching contacts for bulk email: %v\n", err)
				return
			}

			fmt.Printf("Starting bulk email to %d contacts\n", len(contacts))

			for i, contact := range contacts {
				// Create a copy of the request for this contact
				emailReq := req
				emailReq.RecipientEmail = contact.Email

				// Send email
				if err := s.SendEmail(emailReq); err != nil {
					fmt.Printf("Error sending email to %s: %v\n", contact.Email, err)
				} else {
					fmt.Printf("Email sent to %s (%d/%d)\n", contact.Email, i+1, len(contacts))
				}

				// Wait for 1 minute before next email (except after the last one)
				if i < len(contacts)-1 {
					time.Sleep(1 * time.Minute)
				}
			}
			fmt.Println("Bulk email campaign finished")
		} else {
			// Send single email
			if err := s.SendEmail(req); err != nil {
				fmt.Printf("Error sending email in background: %v\n", err)
			}
		}
	}()
	return nil
}
