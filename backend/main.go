package main

import (
	"log"
	"property-brochure-backend/config"
	"property-brochure-backend/handlers"
	"property-brochure-backend/middleware"
	"property-brochure-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate required environment variables
	if cfg.MongoURI == "" {
		log.Fatal("MONGODB_URI is required")
	}
	if cfg.AWSAccessKey == "" || cfg.AWSSecretKey == "" {
		log.Fatal("AWS credentials are required")
	}
	if cfg.AWSS3Bucket == "" {
		log.Fatal("AWS_S3_BUCKET is required")
	}
	if cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	// Initialize services
	log.Println("Connecting to MongoDB...")
	mongoService, err := services.NewMongoDBService(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoService.Close()
	log.Println("Connected to MongoDB successfully")

	log.Println("Initializing AWS S3 service...")
	s3Service, err := services.NewS3Service(
		cfg.AWSAccessKey,
		cfg.AWSSecretKey,
		cfg.AWSRegion,
		cfg.AWSS3Bucket,
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}
	log.Println("AWS S3 service initialized successfully")

	log.Println("Initializing OpenAI service...")
	openaiService := services.NewOpenAIService(cfg.OpenAIAPIKey)
	log.Println("OpenAI service initialized successfully")

	log.Println("Initializing PDF service...")
	pdfService := services.NewPDFService()
	log.Println("PDF service initialized successfully")

	// Initialize handlers
	propertyHandler := handlers.NewPropertyHandler(
		mongoService,
		s3Service,
		openaiService,
		pdfService,
		cfg.MaxFileSize,
		cfg.AllowedFileTypes,
	)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    int(cfg.MaxFileSize * 10), // Allow multiple files
	})

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.SetupCORS(cfg.FrontendURL))

	// Routes
	api := app.Group("/api")
	
	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"message": "Property Brochure API is running",
		})
	})

	// Property endpoints
	api.Post("/property", propertyHandler.SubmitProperty)

	// Start server
	log.Printf("Server starting on port %s...", cfg.Port)
	log.Printf("CORS enabled for: %s", cfg.FrontendURL)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

