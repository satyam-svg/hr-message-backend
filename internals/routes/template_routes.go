package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr_backend/internals/handlers"
	"github.com/satyam-svg/hr_backend/internals/middleware"
)

func SetupTemplateRoutes(app *fiber.App, templateHandler *handlers.TemplateHandler) {
	template := app.Group("/api/template")

	// Protected routes - require authentication
	template.Get("/", middleware.AuthRequired(), templateHandler.GetTemplate)
	template.Put("/", middleware.AuthRequired(), templateHandler.UpdateTemplate)
	template.Post("/enhance", middleware.AuthRequired(), templateHandler.EnhanceTemplate)
}
