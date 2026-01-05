package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
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

// SendEmail sends an email using Gmail SMTP with custom dialer to avoid timeouts
func (s *EmailService) SendEmail(req models.SendEmailRequest) error {
	// Create a new message using gomail for proper MIME/Attachment handling
	m := gomail.NewMessage()
	m.SetHeader("From", req.SenderEmail)
	m.SetHeader("To", req.RecipientEmail)
	m.SetHeader("Subject", req.Subject)

	// Convert newlines to HTML breaks
	htmlBody := strings.ReplaceAll(req.Body, "\n", "<br>")
	m.SetBody("text/html", htmlBody)

	for _, filePath := range req.AttachmentPaths {
		m.Attach(filePath)
	}

	// Configure SMTP Settings
	smtpHost := "smtp.gmail.com"
	smtpPort := "465" // SMTPS (Implicit SSL)

	// Custom Dialer with Timeout (Critical for Render)
	d := net.Dialer{
		Timeout: 30 * time.Second,
	}

	// 1. Connect to SMTP Server using TCP4 (IPv4 Only) and Implicit TLS
	// Use tls.DialWithDialer to combine custom dialer (timeout) with SSL
	tlsConfig := &tls.Config{
		ServerName: smtpHost,
	}

	conn, err := tls.DialWithDialer(&d, "tcp4", smtpHost+":"+smtpPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial smtp (SMTPS/465): %w", err)
	}
	defer conn.Close()

	// 2. Initialize SMTP Client (already secure)
	c, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer c.Quit()

	// 3. Authenticate
	auth := smtp.PlainAuth(
		"",
		req.SenderEmail,
		req.SenderPassword,
		smtpHost,
	)
	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// 4. Send Email
	if err = c.Mail(req.SenderEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = c.Rcpt(req.RecipientEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// 5. Write Data (using gomail's WriteTo to preserve MIME structure)
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	if _, err = m.WriteTo(w); err != nil {
		w.Close() // Close writer if write fails
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// SendEmailForUser sends an email synchronously, fetching credentials from DB
func (s *EmailService) SendEmailForUser(userId string, req models.SendEmailRequest) error {
	ctx := context.Background()

	// 1. Fetch User Credentials
	user, err := s.client.User.FindUnique(
		db.User.ID.Equals(userId),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	professionalEmail, ok1 := user.ProfessionalEmail()
	mailAppPassword, ok2 := user.MailAppPassword()
	if !ok1 || !ok2 || professionalEmail == "" || mailAppPassword == "" {
		return fmt.Errorf("user email credentials not configured. Please configure them in settings")
	}

	// 2. Set Credentials in Request
	req.SenderEmail = professionalEmail
	req.SenderPassword = mailAppPassword

	// 3. Send Logic
	if req.SendToAll {
		// Fetch all contacts (reuse existing logic from StartEmailCampaign or just basic fetch)
		contacts, err := s.client.Contact.FindMany(
			db.Contact.UserID.Equals(userId),
		).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch contacts: %w", err)
		}

		// fmt.Printf("Starting synchronous bulk email to %d contacts\n", len(contacts))

		for i, contact := range contacts {
			emailReq := req
			emailReq.RecipientEmail = contact.Email

			// Add basic personalization if simple body
			// (StartEmailCampaign does this better with templates, but respecting existing structure)

			if err := s.SendEmail(emailReq); err != nil {
				fmt.Printf("Error sending email to %s: %v\n", contact.Email, err)
				// Don't abort entire batch on single failure? Or return error?
				// For synchronous HTTP response, maybe we should return error if critical,
				// but for bulk, usually we want to try best effort.
				// Returning error here will stop the loop and return 500 to user.
				// Let's log and continue for now, or maybe return error?
				// User said: "return error" in "Fix #1".
				return fmt.Errorf("failed to send to %s: %w", contact.Email, err)
			}

			// Small delay to be nice to Gmail
			if i < len(contacts)-1 {
				time.Sleep(2 * time.Second) // 2 seconds delay
			}
		}
	} else {
		// Single Email
		if err := s.SendEmail(req); err != nil {
			return err
		}
	}

	return nil
}
