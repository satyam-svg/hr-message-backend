package services

import (
	"context"
	"fmt"

	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/prisma/db"
)

// TemplateService handles template business logic
type TemplateService struct {
	client *db.PrismaClient
}

// NewTemplateService creates a new template service
func NewTemplateService(client *db.PrismaClient) *TemplateService {
	return &TemplateService{
		client: client,
	}
}

// GetTemplate retrieves the user's template
func (s *TemplateService) GetTemplate(ctx context.Context, userID string) (*models.TemplateResponse, error) {
	template, err := s.client.Template.FindUnique(
		db.Template.UserID.Equals(userID),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}

	return &models.TemplateResponse{
		ID:        template.ID,
		Name:      template.Name,
		Subject:   template.Subject,
		Body:      template.Body,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
	}, nil
}

// UpdateTemplate updates the user's template
func (s *TemplateService) UpdateTemplate(ctx context.Context, userID string, req models.UpdateTemplateRequest) (*models.TemplateResponse, error) {
	// Try to find existing template
	template, err := s.client.Template.FindUnique(
		db.Template.UserID.Equals(userID),
	).Update(
		db.Template.Name.Set(req.Name),
		db.Template.Subject.Set(req.Subject),
		db.Template.Body.Set(req.Body),
	).Exec(ctx)

	// If template doesn't exist (e.g. old user), create one
	if err != nil {
		template, err = s.client.Template.CreateOne(
			db.Template.Name.Set(req.Name),
			db.Template.Subject.Set(req.Subject),
			db.Template.Body.Set(req.Body),
			db.Template.User.Link(db.User.ID.Equals(userID)),
		).Exec(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to update/create template: %w", err)
		}
	}

	return &models.TemplateResponse{
		ID:        template.ID,
		Name:      template.Name,
		Subject:   template.Subject,
		Body:      template.Body,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
	}, nil
}
