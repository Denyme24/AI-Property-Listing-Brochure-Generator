package middleware

import (
	"log"
	"property-brochure-backend/models"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log the error
	log.Printf("Error: %v", err)

	// Return JSON error response
	return c.Status(code).JSON(models.ErrorResponse{
		Success: false,
		Message: message,
		Error:   err.Error(),
	})
}

