package handlers

import (
	"context"
	"fmt"
	"log"
	"property-brochure-backend/models"
	"property-brochure-backend/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PropertyHandler struct {
	mongoService  *services.MongoDBService
	s3Service     *services.S3Service
	openaiService *services.OpenAIService
	pdfService    *services.PDFService
	maxFileSize   int64
	allowedTypes  string
}

func NewPropertyHandler(
	mongo *services.MongoDBService,
	s3 *services.S3Service,
	openai *services.OpenAIService,
	pdf *services.PDFService,
	maxFileSize int64,
	allowedTypes string,
) *PropertyHandler {
	return &PropertyHandler{
		mongoService:  mongo,
		s3Service:     s3,
		openaiService: openai,
		pdfService:    pdf,
		maxFileSize:   maxFileSize,
		allowedTypes:  allowedTypes,
	}
}

func (h *PropertyHandler) SubmitProperty(c *fiber.Ctx) error {
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
	}

	// Extract form values
	req := models.PropertyRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Currency:    c.FormValue("currency", "Dollar"),
		Address:     c.FormValue("address"),
		City:        c.FormValue("city"),
		State:       c.FormValue("state"),
		ZipCode:     c.FormValue("zipCode"),
		AgentName:   c.FormValue("agentName"),
		AgentEmail:  c.FormValue("agentEmail"),
		AgentPhone:  c.FormValue("agentPhone"),
	}

	// Parse price
	if _, err := fmt.Sscanf(c.FormValue("price"), "%f", &req.Price); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid price format",
			Error:   err.Error(),
		})
	}

	// Get amenities
	if amenities, ok := form.Value["amenities[]"]; ok {
		req.Amenities = amenities
	}

	// Validate required fields
	if err := h.validateRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Upload images to S3
	imageURLs := []string{}
	if images, ok := form.File["images[]"]; ok {
		for _, fileHeader := range images {
			// Validate file size
			if fileHeader.Size > h.maxFileSize {
				return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
					Success: false,
					Message: "File size exceeds maximum allowed size",
					Error:   fmt.Sprintf("File %s is too large", fileHeader.Filename),
				})
			}

			// Validate file type
			if !h.isAllowedFileType(fileHeader.Header.Get("Content-Type")) {
				return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
					Success: false,
					Message: "Invalid file type",
					Error:   fmt.Sprintf("File %s has invalid type", fileHeader.Filename),
				})
			}

			// Open file
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening file: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
					Success: false,
					Message: "Failed to process image",
					Error:   err.Error(),
				})
			}
			defer file.Close()

			// Upload to S3
			url, err := h.s3Service.UploadFile(file, fileHeader, "properties")
			if err != nil {
				log.Printf("Error uploading to S3: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
					Success: false,
					Message: "Failed to upload image",
					Error:   err.Error(),
				})
			}

			imageURLs = append(imageURLs, url)
		}
	}

	// Generate AI content
	log.Println("Generating AI content...")
	aiContent, err := h.openaiService.GeneratePropertyContent(
		req.Title,
		req.Description,
		fmt.Sprintf("%.2f", req.Price),
		req.Currency,
		req.Amenities,
	)
	if err != nil {
		log.Printf("Error generating AI content: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to generate AI content",
			Error:   err.Error(),
		})
	}

	// Create property document
	property := &models.Property{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Currency:    req.Currency,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		ZipCode:     req.ZipCode,
		Amenities:   req.Amenities,
		ImageURLs:   imageURLs,
		AgentInfo: models.AgentInfo{
			Name:  req.AgentName,
			Email: req.AgentEmail,
			Phone: req.AgentPhone,
		},
		AIContent: models.AIContent{
			EnglishDescription: aiContent.EnglishDescription,
			ArabicDescription:  aiContent.ArabicDescription,
			KeyHighlights:      aiContent.KeyHighlights,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate PDF brochure
	log.Println("Generating PDF brochure...")
	pdfData, err := h.pdfService.GenerateBrochure(property)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to generate PDF",
			Error:   err.Error(),
		})
	}

	// Upload PDF to S3
	log.Println("Uploading PDF to S3...")
	pdfURL, err := h.s3Service.UploadPDF(pdfData, property.Title)
	if err != nil {
		log.Printf("Error uploading PDF: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to upload PDF",
			Error:   err.Error(),
		})
	}

	property.PDFUrl = pdfURL

	// Save to MongoDB
	log.Println("Saving to MongoDB...")
	collection := h.mongoService.GetCollection("properties")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, property)
	if err != nil {
		log.Printf("Error saving to MongoDB: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to save property",
			Error:   err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(models.PropertyResponse{
		Success:    true,
		Message:    "Property listing created successfully",
		PropertyID: property.ID.Hex(),
		PDFUrl:     pdfURL,
	})
}

func (h *PropertyHandler) validateRequest(req *models.PropertyRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if req.Address == "" {
		return fmt.Errorf("address is required")
	}
	if req.City == "" {
		return fmt.Errorf("city is required")
	}
	if req.State == "" {
		return fmt.Errorf("state is required")
	}
	if req.ZipCode == "" {
		return fmt.Errorf("zip code is required")
	}
	if req.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	if req.AgentEmail == "" {
		return fmt.Errorf("agent email is required")
	}
	if req.AgentPhone == "" {
		return fmt.Errorf("agent phone is required")
	}
	return nil
}

func (h *PropertyHandler) isAllowedFileType(contentType string) bool {
	allowedTypes := strings.Split(h.allowedTypes, ",")
	for _, allowed := range allowedTypes {
		if strings.TrimSpace(allowed) == contentType {
			return true
		}
	}
	return false
}

