package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr_backend/internals/handlers"
	"github.com/satyam-svg/hr_backend/internals/middleware"
)

// SetupEmailRoutes sets up email routes
func SetupEmailRoutes(app *fiber.App, emailHandler *handlers.EmailHandler) {
	// Create email group
	email := app.Group("/api/email")

	// Protected routes
	email.Post("/send", middleware.AuthRequired(), emailHandler.SendEmail)
}
