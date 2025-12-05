package services

import (
	"context"
	"fmt"

	"github.com/satyam-svg/hr-message-backend/internals/models"
	"github.com/satyam-svg/hr-message-backend/prisma/db"
)

type ContactService struct {
	client *db.PrismaClient
}

func NewContactServcie(client *db.PrismaClient) *ContactService {
	return &ContactService{
		client: client,
	}
}

func (s *ContactService) UpdateContact(ctx context.Context, contactId string, userId string, req models.UpdateContactRequest) (*models.ContactResponse, error) {
	// Verify ownership
	_, err := s.client.Contact.FindFirst(
		db.Contact.ID.Equals(contactId),
		db.Contact.UserID.Equals(userId),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("contact not found or unauthorized")
	}

	var params []db.ContactSetParam
	if req.Name != nil {
		params = append(params, db.Contact.Name.Set(*req.Name))
	}
	if req.CompanyName != nil {
		params = append(params, db.Contact.CompanyName.Set(*req.CompanyName))
	}
	if req.Email != nil {
		params = append(params, db.Contact.Email.Set(*req.Email))
	}
	if req.IsSent != nil {
		params = append(params, db.Contact.IsSent.Set(*req.IsSent))
	}

	updated, err := s.client.Contact.FindUnique(
		db.Contact.ID.Equals(contactId),
	).Update(
		params...,
	).Exec(ctx)

	if err != nil {
		return nil, err
	}

	return &models.ContactResponse{
		ID:          updated.ID,
		Name:        updated.Name,
		CompanyName: updated.CompanyName,
		Email:       updated.Email,
		IsSent:      updated.IsSent,
		CreatedAt:   updated.CreatedAt,
	}, nil
}
