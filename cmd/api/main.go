package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/satyam-svg/hr_backend/internals/handlers"
	"github.com/satyam-svg/hr_backend/internals/routes"
	"github.com/satyam-svg/hr_backend/internals/services"
	"github.com/satyam-svg/hr_backend/prisma/db"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Prisma client
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			log.Printf("Failed to disconnect from database: %v", err)
		}
	}()

	log.Println("✅ Connected to database")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   "server_error",
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	// Initialize services
	authService := services.NewAuthService(client)
	emailService := services.NewEmailService(client)
	templateService := services.NewTemplateService(client)
	contactService := services.NewContactServcie(client)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	emailHandler := handlers.NewEmailHandler(emailService)
	pdfHandler := handlers.NewPDFHandler(client)
	templateHandler := handlers.NewTemplateHandler(templateService)
	contactHandler := handlers.NewContactHandler(contactService)

	// Setup routes
	routes.SetupAuthRoutes(app, authHandler)
	routes.SetupEmailRoutes(app, emailHandler)
	routes.SetupPDFRoutes(app, pdfHandler)
	routes.SetupTemplateRoutes(app, templateHandler)
	routes.SetupContactRoutes(app, contactHandler)

	// Health check endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "HR Backend API is running",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database connection
		ctx := context.Background()
		_, err := client.User.FindMany().Exec(ctx)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "error",
				"database": "disconnected",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server is running on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("❌ Fiber server stopped: %v", err)
		}
	}()

	<-quit

	log.Println("⚠️  Shutdown initiated...")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(); err != nil {
		log.Printf("❌ Fiber shutdown error: %v", err)
	}

	// Gracefully disconnect Prisma
	if err := client.Prisma.Disconnect(); err != nil {
		log.Printf("❌ Failed to disconnect Prisma: %v", err)
	}

	log.Println("✅ Fiber server shutdown successfully")

}
