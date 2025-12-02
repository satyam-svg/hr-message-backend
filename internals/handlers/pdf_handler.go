package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/satyam-svg/hr_backend/prisma/db"
	"google.golang.org/api/option"
)

// PDFHandler handles PDF HTTP requests
type PDFHandler struct {
	client *db.PrismaClient
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(client *db.PrismaClient) *PDFHandler {
	return &PDFHandler{
		client: client,
	}
}

func (h *PDFHandler) UploadPDF(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	// Ensure uploads directory exists
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create uploads directory"})
	}

	// Save file temporarily
	filePath := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer os.Remove(filePath) // Clean up after processing

	// Read file content
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to read file"})
	}

	// Initialize Gemini Client
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(500).JSON(fiber.Map{"error": "GEMINI_API_KEY not set"})
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create gemini client"})
	}
	defer client.Close()

	// Configure Model
	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := []genai.Part{
		genai.Blob{
			MIMEType: "application/pdf",
			Data:     fileBytes,
		},
		genai.Text(`Extract the following details for the first 20 companies found in this PDF: 
		- Company Name
		- Email
		- Name
		
		Return the result as a JSON object with the following structure:
		{
			"companies": [
				{
			        "name":"HR name"
			        "company_name": "Company Name",
					"email": "Email Address"
				}
			]
		}
		`),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate content from gemini"})
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return c.Status(500).JSON(fiber.Map{"error": "no content generated"})
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
		Companies []struct {
			Name        string `json:"name"`
			CompanyName string `json:"company_name"`
			Email       string `json:"email"`
		} `json:"companies"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("Failed to parse JSON directly: %v. Raw: %s", err, jsonStr)
		return c.Status(500).JSON(fiber.Map{"error": "failed to parse gemini response", "raw": jsonStr})
	}

	// Save contacts to database
	savedContacts := []map[string]interface{}{}
	for _, company := range result.Companies {
		contact, err := h.client.Contact.CreateOne(
			db.Contact.Name.Set(company.Name),
			db.Contact.CompanyName.Set(company.CompanyName),
			db.Contact.Email.Set(company.Email),
			db.Contact.User.Link(db.User.ID.Equals(userID)),
		).Exec(ctx)

		if err != nil {
			log.Printf("Failed to save contact: %v", err)
			continue
		}

		savedContacts = append(savedContacts, map[string]interface{}{
			"id":           contact.ID,
			"name":         contact.Name,
			"company_name": contact.CompanyName,
			"email":        contact.Email,
			"is_sent":      contact.IsSent,
		})
	}

	return c.JSON(fiber.Map{
		"message":        "PDF processed successfully",
		"contacts_saved": len(savedContacts),
		"contacts":       savedContacts,
	})
}
