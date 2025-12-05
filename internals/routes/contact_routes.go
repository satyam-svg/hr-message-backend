package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr-message-backend/internals/handlers"
	"github.com/satyam-svg/hr-message-backend/internals/middleware"
)

func SetupContactRoutes(app *fiber.App, contactHandler *handlers.ContactHandler) {
	api := app.Group("/api")

	// Protected routes
	contacts := api.Group("/contacts", middleware.AuthRequired())
	contacts.Patch("/:id", contactHandler.UpdateContact)
}
