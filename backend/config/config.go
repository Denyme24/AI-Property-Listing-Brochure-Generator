package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	FrontendURL       string
	MongoURI          string
	MongoDatabase     string
	AWSAccessKey      string
	AWSSecretKey      string
	AWSRegion         string
	AWSS3Bucket       string
	OpenAIAPIKey      string
	MaxFileSize       int64
	AllowedFileTypes  string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	maxFileSize, err := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64)
	if err != nil {
		maxFileSize = 10485760 // Default 10MB
	}

	return &Config{
		Port:              getEnv("PORT", "8000"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		MongoURI:          getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:     getEnv("MONGODB_DATABASE", "property_brochure_db"),
		AWSAccessKey:      getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:         getEnv("AWS_REGION", "us-east-1"),
		AWSS3Bucket:       getEnv("AWS_S3_BUCKET", ""),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		MaxFileSize:       maxFileSize,
		AllowedFileTypes:  getEnv("ALLOWED_FILE_TYPES", "image/jpeg,image/jpg,image/png,image/webp"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

