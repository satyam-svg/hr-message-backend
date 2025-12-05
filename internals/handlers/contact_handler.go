package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satyam-svg/hr_backend/internals/models"
	"github.com/satyam-svg/hr_backend/internals/services"
)

type ContactHandler struct {
	service *services.ContactService
}

func NewContactHandler(service *services.ContactService) *ContactHandler {
	return &ContactHandler{
		service: service,
	}
}

func (h *ContactHandler) UpdateContact(c *fiber.Ctx) error {
	contactId := c.Params("id")
	if contactId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "contact id is required"})
	}

	var req models.UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	userId := c.Locals("userId").(string)

	updatedContact, err := h.service.UpdateContact(c.Context(), contactId, userId, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(updatedContact)
}
