package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/internals/services"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Signup handles user registration
// POST /api/auth/signup
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var req models.SignupRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Email, password, and name are required",
		})
	}

	// Validate password length
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Password must be at least 8 characters long",
		})
	}

	// Call service
	response, err := h.authService.Signup(c.Context(), req)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
				Error:   "user_exists",
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// Login handles user authentication
// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Email and password are required",
		})
	}

	// Call service
	response, err := h.authService.Login(c.Context(), req)
	if err != nil {
		if err.Error() == "invalid email or password" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "invalid_credentials",
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to authenticate user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateEmailSettings handles updating professional email and mail app password
// PUT /api/auth/email-settings
func (h *AuthHandler) UpdateEmailSettings(c *fiber.Ctx) error {
	var req models.UpdateEmailSettingsRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.ProfessionalEmail == "" || req.MailAppPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: "Professional email and mail app password are required",
		})
	}

	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	// Call service
	if err := h.authService.UpdateEmailSettings(c.Context(), userID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update email settings",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email settings updated successfully",
	})
}

// GetProfile handles fetching the authenticated user's profile
// GET /api/auth/me
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID := c.Locals("userId").(string)

	profile, err := h.authService.GetProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to fetch user profile- " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}
