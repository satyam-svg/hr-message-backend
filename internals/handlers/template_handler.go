package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/internals/services"
	"google.golang.org/api/option"
)

// TemplateHandler handles template HTTP requests
type TemplateHandler struct {
	templateService *services.TemplateService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// GetTemplate handles fetching the user's template
// GET /api/template
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	template, err := h.templateService.GetTemplate(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "not_found",
			Message: "Template not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(template)
}

// UpdateTemplate handles updating the user's template
// PUT /api/template
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	var req models.UpdateTemplateRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.Subject == "" || req.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Name, subject, and body are required",
		})
	}

	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	template, err := h.templateService.UpdateTemplate(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update template",
		})
	}

	return c.Status(fiber.StatusOK).JSON(template)
}

// EnhanceTemplate handles enhancing the template content using Gemini
// POST /api/template/enhance
func (h *TemplateHandler) EnhanceTemplate(c *fiber.Ctx) error {
	var req models.EnhanceTemplateRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Content is required",
		})
	}

	// Initialize Gemini Client
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "GEMINI_API_KEY not set",
		})
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create gemini client",
		})
	}
	defer client.Close()

	// Configure Model
	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`Enhance the following email template content to make it sound professional, persuasive, and human-written.
	- Improve clarity, flow, and professionalism.
	- Use a natural, polite, and confident tone.
	- Do NOT use markdown formatting (like ** or ##).
	- Use HTML <b> tags sparingly for emphasis on key achievements or important keywords.
	- ALWAYS use HTML <b> tags for technical skills, programming languages, frameworks, and tools (e.g. <b>React</b>, <b>Go</b>, <b>AWS</b>).
	- Use HTML <br> tags for line breaks to ensuring proper formatting.
	- Ensure placeholders like {Company}, {Position}, {Hiring Manager Name} etc., are preserved exactly as is.
	- The output should be valid HTML ready to be embedded in an email body.

	Content:
	%s

	Return the result as a JSON object with the following structure:
	{
		"enhanced_content": "The enhanced text here"
	}`, req.Content)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to enhance content",
		})
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "No content generated",
		})
	}

	// Extract JSON string
	var jsonStr string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			jsonStr += string(txt)
		}
	}

	// Parse JSON
	var result struct {
		EnhancedContent string `json:"enhanced_content"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("Failed to parse JSON directly: %v. Raw: %s", err, jsonStr)
		// Fallback: use raw text if JSON parsing fails
		result.EnhancedContent = jsonStr
	}

	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	// Update the template in the database with the enhanced content
	// We assume the user wants to update the 'Body' of the template.
	// If the request also included Subject, we could enhance that too, but for now let's stick to Body/Content.
	updatedTemplate, err := h.templateService.UpdateTemplate(c.Context(), userID, models.UpdateTemplateRequest{
		// We need to fetch the existing template first to preserve Name and Subject if we are only updating Body
		// But for simplicity, let's fetch it first.
		Body: result.EnhancedContent,
	})

	// Fetch existing template to preserve other fields
	existingTemplate, err := h.templateService.GetTemplate(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to fetch existing template for update",
		})
	}

	// Update with enhanced body, keeping existing name and subject
	updatedTemplate, err = h.templateService.UpdateTemplate(c.Context(), userID, models.UpdateTemplateRequest{
		Name:    existingTemplate.Name,
		Subject: existingTemplate.Subject,
		Body:    result.EnhancedContent,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to save enhanced template",
		})
	}

	return c.Status(fiber.StatusOK).JSON(updatedTemplate)
}
