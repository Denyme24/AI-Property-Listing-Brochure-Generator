package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Property represents a property listing in the database
type Property struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Currency    string             `bson:"currency" json:"currency"`
	Address     string             `bson:"address" json:"address"`
	City        string             `bson:"city" json:"city"`
	State       string             `bson:"state" json:"state"`
	ZipCode     string             `bson:"zipCode" json:"zipCode"`
	Amenities   []string           `bson:"amenities" json:"amenities"`
	ImageURLs   []string           `bson:"imageUrls" json:"imageUrls"`
	AgentInfo   AgentInfo          `bson:"agentInfo" json:"agentInfo"`
	AIContent   AIContent          `bson:"aiContent" json:"aiContent"`
	PDFUrl      string             `bson:"pdfUrl" json:"pdfUrl"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// AgentInfo represents the real estate agent's contact information
type AgentInfo struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Phone string `bson:"phone" json:"phone"`
}

// AIContent represents AI-generated content for the property
type AIContent struct {
	EnglishDescription string   `bson:"englishDescription" json:"englishDescription"`
	ArabicDescription  string   `bson:"arabicDescription" json:"arabicDescription"`
	KeyHighlights      []string `bson:"keyHighlights" json:"keyHighlights"`
}

// PropertyRequest represents the incoming request data
type PropertyRequest struct {
	Title       string   `form:"title" validate:"required"`
	Description string   `form:"description"`
	Price       float64  `form:"price" validate:"required"`
	Currency    string   `form:"currency"`
	Address     string   `form:"address" validate:"required"`
	City        string   `form:"city" validate:"required"`
	State       string   `form:"state" validate:"required"`
	ZipCode     string   `form:"zipCode" validate:"required"`
	Amenities   []string `form:"amenities[]"`
	AgentName   string   `form:"agentName" validate:"required"`
	AgentEmail  string   `form:"agentEmail" validate:"required,email"`
	AgentPhone  string   `form:"agentPhone" validate:"required"`
}

// PropertyResponse represents the API response
type PropertyResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	PropertyID     string `json:"propertyId,omitempty"`
	PDFUrl         string `json:"pdfUrl,omitempty"`
	PDFViewUrl     string `json:"pdfViewUrl,omitempty"`
	PDFDownloadUrl string `json:"pdfDownloadUrl,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

