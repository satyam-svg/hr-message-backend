package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/internals/services"
)

// EmailHandler handles email HTTP requests
type EmailHandler struct {
	emailService *services.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// StartCampaign handles starting an email campaign
// POST /api/email/campaign/start
func (h *EmailHandler) StartCampaign(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)

	if err := h.emailService.StartEmailCampaign(userId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "campaign_error",
			Message: "Failed to start campaign: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Email campaign started in background",
		"success": true,
	})
}

// SendEmail handles sending an email
// POST /api/email/send
func (h *EmailHandler) SendEmail(c *fiber.Ctx) error {
	var req models.SendEmailRequest
	var attachmentPaths []string

	// Check content type
	contentType := c.Get("Content-Type")

	// Handle multipart/form-data (with file uploads)
	if len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Parse form fields
		req.SendToAll = c.FormValue("send_to_all") == "true"
		req.SenderEmail = c.FormValue("sender_email")
		// Password is no longer accepted from client for security
		if !req.SendToAll {
			req.RecipientEmail = c.FormValue("recipient_email")
		}
		req.Subject = c.FormValue("subject")
		req.Body = c.FormValue("body")

		// Handle file uploads
		form, err := c.MultipartForm()
		if err == nil && form.File["attachments"] != nil {
			files := form.File["attachments"]

			// Create uploads directory if it doesn't exist
			uploadsDir := "./uploads/temp"
			if err := os.MkdirAll(uploadsDir, 0755); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
					Error:   "directory_creation_error",
					Message: "Failed to create upload directory: " + err.Error(),
				})
			}

			for _, file := range files {
				// Save file temporarily
				filePath := uploadsDir + "/" + file.Filename
				if err := c.SaveFile(file, filePath); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
						Error:   "file_upload_error",
						Message: "Failed to save uploaded file: " + err.Error(),
					})
				}
				attachmentPaths = append(attachmentPaths, filePath)
			}
		}

		req.AttachmentPaths = attachmentPaths
	} else {
		// Handle JSON request (backward compatibility)
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid request body",
			})
		}
	}

	// Validate required fields
	// Note: sender_email and sender_password are no longer required from client as they are fetched from DB
	if req.Subject == "" || req.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Required fields missing: subject, body",
		})
	}

	if !req.SendToAll && req.RecipientEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "recipient_email is required when send_to_all is false",
		})
	}

	// Get UserID from context
	userId := c.Locals("userId").(string)

	// Call service to send email synchronously (Foreground)
	// This helps catch errors immediately and prevents Render from killing background goroutines
	if err := h.emailService.SendEmailForUser(userId, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "email_send_error",
			Message: "Failed to send email: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SendEmailResponse{
		Message: "Email sent successfully",
		Success: true,
	})
}
