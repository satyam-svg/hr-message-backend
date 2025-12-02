package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr_backend/internals/handlers"
	"github.com/satyam-svg/hr_backend/internals/middleware"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(app *fiber.App, authHandler *handlers.AuthHandler) {
	// Create auth group
	auth := app.Group("/api/auth")

	// Public routes
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	auth.Put("/email-settings", middleware.AuthRequired(), authHandler.UpdateEmailSettings)
	auth.Get("/me", middleware.AuthRequired(), authHandler.GetProfile)
}
