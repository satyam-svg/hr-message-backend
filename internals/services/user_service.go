package services

import (
	"context"
	"fmt"
	"time"

	"github.com/satyam-svg/hr-message-backend/prisma/db"
)

type UserService struct {
	client *db.PrismaClient
}

func NewUserService(client *db.PrismaClient) *UserService {
	return &UserService{
		client: client,
	}
}

// IncrementPDFUploadCount increments the pdfUploadCount for a user and logs an activity
func (s *UserService) IncrementPDFUploadCount(userId string) error {
	ctx := context.Background()
	// Increment count and create activity sequentially

	// 1. Increment count
	_, err := s.client.User.FindUnique(
		db.User.ID.Equals(userId),
	).Update(
		db.User.PdfUploadCount.Increment(1),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment pdf upload count: %w", err)
	}

	// 2. Create Activity
	_, err = s.client.Activity.CreateOne(
		db.Activity.Description.Set("Uploaded a PDF"),
		db.Activity.User.Link(db.User.ID.Equals(userId)),
		db.Activity.CreatedAt.Set(time.Now()),
	).Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	return nil
}
