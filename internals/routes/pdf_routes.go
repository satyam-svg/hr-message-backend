package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr-message-backend/internals/handlers"
	"github.com/satyam-svg/hr-message-backend/internals/middleware"
)

func SetupPDFRoutes(app *fiber.App, pdfHandler *handlers.PDFHandler) {
	// Protected route - requires authentication
	app.Post("/upload", middleware.AuthRequired(), pdfHandler.UploadPDF)
}
