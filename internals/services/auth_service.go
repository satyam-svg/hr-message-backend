package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/satyam-svg/hr_backend/internals/models"
	"github.com/satyam-svg/hr_backend/internals/utils"
	"github.com/satyam-svg/hr_backend/prisma/db"
)

// AuthService handles authentication business logic
type AuthService struct {
	client *db.PrismaClient
}

// NewAuthService creates a new auth service
func NewAuthService(client *db.PrismaClient) *AuthService {
	return &AuthService{
		client: client,
	}
}

// Signup registers a new user with concurrent password hashing
func (s *AuthService) Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error) {
	// Channel for concurrent operations
	type result struct {
		data interface{}
		err  error
	}

	// Check if user already exists
	existingUser, err := s.client.User.FindUnique(
		db.User.Email.Equals(req.Email),
	).Exec(ctx)

	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password concurrently
	hashChan := make(chan result)
	go func() {
		hash, err := utils.HashPassword(req.Password)
		hashChan <- result{data: hash, err: err}
	}()

	// Wait for password hashing
	hashResult := <-hashChan
	if hashResult.err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", hashResult.err)
	}

	hashedPassword := hashResult.data.(string)

	// Create user
	user, err := s.client.User.CreateOne(
		db.User.Name.Set(req.Name),
		db.User.Email.Set(req.Email),
		db.User.Password.Set(hashedPassword),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default email template
	defaultSubject := "Application for {Position} at {Company}"
	defaultBody := `Dear {Hiring Manager Name},

I hope you are doing well. My name is Praveen Maurya, and I am currently in my final year at IIIT Ranchi, pursuing B.Tech in Electronics and Communication Engineering. I am writing to express my interest in the {Position} role at {Company}.

I bring strong full-stack development experience, having built high-scale systems using Go, React/Next.js, Node.js, PostgreSQL, Redis, and cloud services like AWS (EC2, RDS, S3). I have also developed multiple 3D/Interactive applications using Three.js, WebGL, and advanced animation systems, which I believe aligns well with companies working on modern web technologies.

In my recent roles at Samsara Innovations and Alpixn Technologies, I have:

Built Go microservices, 3D visualization dashboards, and optimized APIs by up to 65%

Developed 30+ production-ready screens in React Native and React.js

Created 3 SaaS applications with OAuth, Redis caching, and scalable database architecture

Designed large ETL pipelines (handling 2TB+/month) with secure cloud deployments

Apart from work experience, I have built impactful projects such as:

Inclusify (E-commerce + Blockchain verification + 3D product previews)

AI Document Scanner (92% accuracy)

SalonSphere (10K+ daily transactions, <500ms latency)

Interactive 3D Portfolio using Three.js

I am confident that my experience in full-stack engineering, distributed systems, cloud deployment, and 3D interactive development will allow me to contribute meaningfully to your team.

I have attached my resume for your review. I would be grateful for the opportunity to discuss how I can add value at {Company}.

Thank you for considering my application. I look forward to hearing from you.

Best regards,
Praveen Maurya
Email: satyammaurya9620@gmail.com

Portfolio: https://iampraveen.vercel.app

LinkedIn: https://www.linkedin.com/in/praveenmaurya07

GitHub: https://github.com/satyam-svg`

	_, err = s.client.Template.CreateOne(
		db.Template.Name.Set("Default Template"),
		db.Template.Subject.Set(defaultSubject),
		db.Template.Body.Set(defaultBody),
		db.Template.User.Link(db.User.ID.Equals(user.ID)),
	).Exec(ctx)

	if err != nil {
		// Log error but don't fail signup (optional: could delete user and fail)
		fmt.Printf("Failed to create default template: %v\n", err)
	}

	// Generate JWT token concurrently
	tokenChan := make(chan result)
	go func() {
		token, err := utils.GenerateToken(user.ID, user.Email)
		tokenChan <- result{data: token, err: err}
	}()

	// Wait for token generation
	tokenResult := <-tokenChan
	if tokenResult.err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", tokenResult.err)
	}

	token := tokenResult.data.(string)

	// Prepare response
	response := &models.AuthResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}

	return response, nil
}

// Login authenticates a user with concurrent password verification
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	// Channel for concurrent operations
	type result struct {
		data interface{}
		err  error
	}

	// Find user by email
	user, err := s.client.User.FindUnique(
		db.User.Email.Equals(req.Email),
	).Exec(ctx)

	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password concurrently
	passwordChan := make(chan error)
	go func() {
		err := utils.ComparePassword(user.Password, req.Password)
		passwordChan <- err
	}()

	// Wait for password verification
	if err := <-passwordChan; err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token concurrently
	tokenChan := make(chan result)
	go func() {
		token, err := utils.GenerateToken(user.ID, user.Email)
		tokenChan <- result{data: token, err: err}
	}()

	// Wait for token generation
	tokenResult := <-tokenChan
	if tokenResult.err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", tokenResult.err)
	}

	token := tokenResult.data.(string)

	// Prepare response
	response := &models.AuthResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}

	return response, nil
}

// UpdateEmailSettings updates the user's professional email and mail app password
func (s *AuthService) UpdateEmailSettings(ctx context.Context, userID string, req models.UpdateEmailSettingsRequest) error {
	_, err := s.client.User.FindUnique(
		db.User.ID.Equals(userID),
	).Update(
		db.User.ProfessionalEmail.Set(req.ProfessionalEmail),
		db.User.MailAppPassword.Set(req.MailAppPassword),
	).Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update email settings: %w", err)
	}

	return nil
}

// GetProfile fetches the user's full profile including contacts and template
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
	user, err := s.client.User.FindUnique(
		db.User.ID.Equals(userID),
	).With(
		db.User.Contacts.Fetch(),
		db.User.Template.Fetch(),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	// Map contacts
	contacts := []models.ContactResponse{} // Initialize as empty array
	for _, c := range user.Contacts() {
		contacts = append(contacts, models.ContactResponse{
			ID:          c.ID,
			Name:        c.Name,
			CompanyName: c.CompanyName,
			Email:       c.Email,
			IsSent:      c.IsSent,
			CreatedAt:   c.CreatedAt,
		})
	}

	// Map template
	var template *models.TemplateResponse
	if t, ok := user.Template(); ok {
		template = &models.TemplateResponse{
			ID:        t.ID,
			Name:      t.Name,
			Subject:   t.Subject,
			Body:      t.Body,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		}
	} else {
		// Create default template for existing users if missing
		defaultSubject := "Application for {Position} at {Company}"
		defaultBody := `Dear {Hiring Manager Name},

I hope you are doing well. My name is Praveen Maurya, and I am currently in my final year at IIIT Ranchi, pursuing B.Tech in Electronics and Communication Engineering. I am writing to express my interest in the {Position} role at {Company}.

I bring strong full-stack development experience, having built high-scale systems using Go, React/Next.js, Node.js, PostgreSQL, Redis, and cloud services like AWS (EC2, RDS, S3). I have also developed multiple 3D/Interactive applications using Three.js, WebGL, and advanced animation systems, which I believe aligns well with companies working on modern web technologies.

In my recent roles at Samsara Innovations and Alpixn Technologies, I have:

Built Go microservices, 3D visualization dashboards, and optimized APIs by up to 65%

Developed 30+ production-ready screens in React Native and React.js

Created 3 SaaS applications with OAuth, Redis caching, and scalable database architecture

Designed large ETL pipelines (handling 2TB+/month) with secure cloud deployments

Apart from work experience, I have built impactful projects such as:

Inclusify (E-commerce + Blockchain verification + 3D product previews)

AI Document Scanner (92% accuracy)

SalonSphere (10K+ daily transactions, <500ms latency)

Interactive 3D Portfolio using Three.js

I am confident that my experience in full-stack engineering, distributed systems, cloud deployment, and 3D interactive development will allow me to contribute meaningfully to your team.

I have attached my resume for your review. I would be grateful for the opportunity to discuss how I can add value at {Company}.

Thank you for considering my application. I look forward to hearing from you.

Best regards,
Praveen Maurya
Email: satyammaurya9620@gmail.com

Portfolio: https://iampraveen.vercel.app

LinkedIn: https://www.linkedin.com/in/praveenmaurya07

GitHub: https://github.com/satyam-svg`

		newTemplate, err := s.client.Template.CreateOne(
			db.Template.Name.Set("Default Template"),
			db.Template.Subject.Set(defaultSubject),
			db.Template.Body.Set(defaultBody),
			db.Template.User.Link(db.User.ID.Equals(userID)),
		).Exec(ctx)

		if err == nil {
			template = &models.TemplateResponse{
				ID:        newTemplate.ID,
				Name:      newTemplate.Name,
				Subject:   newTemplate.Subject,
				Body:      newTemplate.Body,
				CreatedAt: newTemplate.CreatedAt,
				UpdatedAt: newTemplate.UpdatedAt,
			}
		}
	}

	professionalEmail := ""
	if v, ok := user.ProfessionalEmail(); ok {
		professionalEmail = v
	}

	mailAppPassword := ""
	if v, ok := user.MailAppPassword(); ok {
		mailAppPassword = v
	}

	return &models.UserProfileResponse{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		ProfessionalEmail: professionalEmail,
		MailAppPassword:   mailAppPassword,
		DailyLimit:        user.DailyLimit,
		CreatedAt:         user.CreatedAt,
		Contacts:          contacts,
		Template:          template,
	}, nil
}
