package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client *openai.Client
}

type AIGeneratedContent struct {
	EnglishDescription string
	ArabicDescription  string
	KeyHighlights      []string
}

type LocalizedContentGenerated struct {
	EnglishContent LocalizedContentData `json:"englishContent"`
	ArabicContent  LocalizedContentData `json:"arabicContent"`
}

type LocalizedContentData struct {
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	Highlights               []string `json:"highlights"`
	TranslatedAmenities      []string `json:"translatedAmenities"`
	PriceLabel               string   `json:"priceLabel"`
	AddressLabel             string   `json:"addressLabel"`
	CityLabel                string   `json:"cityLabel"`
	StateLabel               string   `json:"stateLabel"`
	ZipCodeLabel             string   `json:"zipCodeLabel"`
	AmenitiesLabel           string   `json:"amenitiesLabel"`
	AgentLabel               string   `json:"agentLabel"`
	PropertyDescriptionLabel string   `json:"propertyDescriptionLabel"`
	KeyHighlightsLabel       string   `json:"keyHighlightsLabel"`
	PropertyGalleryLabel     string   `json:"propertyGalleryLabel"`
}

func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		client: openai.NewClient(apiKey),
	}
}

func (s *OpenAIService) GeneratePropertyContent(title, description, price, currency string, amenities []string) (*AIGeneratedContent, error) {
	ctx := context.Background()

	
	englishDesc := description
	if description == "" || len(description) < 50 {
		prompt := fmt.Sprintf(`Generate an engaging and professional property description in English for a real estate listing with the following details:
- Title: %s
- Price: %s %s
- Amenities: %s

The description should be 3-4 paragraphs long, highlight the key features, and appeal to potential buyers. Make it compelling and professional.`, 
			title, price, currency, strings.Join(amenities, ", "))

		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a professional real estate content writer who creates compelling property descriptions.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   500,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate English description: %w", err)
		}
		englishDesc = resp.Choices[0].Message.Content
	}

	// Translate to Arabic
	arabicPrompt := fmt.Sprintf("Translate the following real estate property description to Arabic. Maintain the professional tone and structure:\n\n%s", englishDesc)
	
	arabicResp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a professional translator specializing in real estate content. Translate from English to Arabic while maintaining professionalism.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: arabicPrompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   600,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate Arabic translation: %w", err)
	}
	arabicDesc := arabicResp.Choices[0].Message.Content

	// Generate key highlights
	highlightsPrompt := fmt.Sprintf(`Based on this property listing, generate 5-7 key highlights as short bullet points (each 5-10 words):
Title: %s
Price: %s %s
Amenities: %s
Description: %s

Return only the bullet points, one per line, without bullet symbols or numbering.`, 
		title, price, currency, strings.Join(amenities, ", "), englishDesc)

	highlightsResp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a real estate marketing expert who creates concise, impactful property highlights.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: highlightsPrompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   300,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate highlights: %w", err)
	}

	// Parse highlights
	highlightsText := highlightsResp.Choices[0].Message.Content
	highlights := []string{}
	for _, line := range strings.Split(highlightsText, "\n") {
		line = strings.TrimSpace(line)
		// Remove common bullet point characters
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "• ")
		line = strings.TrimPrefix(line, "* ")
		// Remove numbering
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' {
			parts := strings.SplitN(line, ".", 2)
			if len(parts) == 2 {
				line = strings.TrimSpace(parts[1])
			}
		}
		if line != "" {
			highlights = append(highlights, line)
		}
	}

	return &AIGeneratedContent{
		EnglishDescription: englishDesc,
		ArabicDescription:  arabicDesc,
		KeyHighlights:      highlights,
	}, nil
}

// GenerateLocalizedContent generates fully localized content for both English and Arabic
func (s *OpenAIService) GenerateLocalizedContent(title, description, price, currency string, amenities []string) (*LocalizedContentGenerated, error) {
	ctx := context.Background()

	// Create a comprehensive prompt that asks for both English and Arabic localized content
	prompt := fmt.Sprintf(`You are a professional real estate content generator. Generate fully localized content for a property listing in both English and Arabic.

Property Details:
- Title: %s
- Price: %s %s
- Amenities: %s
- Description: %s

Please generate a JSON response with the following structure:
{
  "englishContent": {
    "title": "<translated/enhanced property title in English>",
    "description": "<3-4 paragraph professional description in English>",
    "highlights": ["<5-7 short key highlights in English, each 5-10 words>"],
    "translatedAmenities": ["<all amenities translated to English>"],
    "priceLabel": "Price",
    "addressLabel": "Address",
    "cityLabel": "City",
    "stateLabel": "State",
    "zipCodeLabel": "ZIP Code",
    "amenitiesLabel": "Amenities & Features",
    "agentLabel": "Contact Your Agent",
    "propertyDescriptionLabel": "Property Description",
    "keyHighlightsLabel": "Key Highlights",
    "propertyGalleryLabel": "Property Gallery"
  },
  "arabicContent": {
    "title": "<property title fully translated to Arabic>",
    "description": "<3-4 paragraph professional description fully in Arabic>",
    "highlights": ["<5-7 short key highlights in Arabic>"],
    "translatedAmenities": ["<all amenities translated to Arabic>"],
    "priceLabel": "السعر",
    "addressLabel": "العنوان",
    "cityLabel": "المدينة",
    "stateLabel": "الولاية",
    "zipCodeLabel": "الرمز البريدي",
    "amenitiesLabel": "المرافق والميزات",
    "agentLabel": "اتصل بوكيلك",
    "propertyDescriptionLabel": "وصف العقار",
    "keyHighlightsLabel": "المميزات الرئيسية",
    "propertyGalleryLabel": "معرض العقار"
  }
}

Important:
1. The Arabic version must be COMPLETELY in Arabic - no English words
2. Translate amenities accurately (e.g., Swimming Pool → حمام السباحة, Parking → موقف سيارات, Garden → حديقة, Gym → صالة رياضية)
3. All labels in Arabic must use proper Arabic terminology
4. Keep highlights concise and impactful
5. Return ONLY valid JSON, no additional text

Generate the content now:`, 
		title, price, currency, strings.Join(amenities, ", "), description)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a professional real estate content generator with expertise in English and Arabic. You always return valid JSON responses.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate localized content: %w", err)
	}

	// Parse the JSON response
	responseText := strings.TrimSpace(resp.Choices[0].Message.Content)
	
	// Remove markdown code blocks if present
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var result LocalizedContentGenerated
	err = json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse localized content JSON: %w\nResponse: %s", err, responseText)
	}

	// Ensure we have all required fields with fallbacks
	if result.EnglishContent.Title == "" {
		result.EnglishContent.Title = title
	}
	if result.EnglishContent.PriceLabel == "" {
		result.EnglishContent.PriceLabel = "Price"
	}
	if result.EnglishContent.AddressLabel == "" {
		result.EnglishContent.AddressLabel = "Address"
	}
	if result.EnglishContent.CityLabel == "" {
		result.EnglishContent.CityLabel = "City"
	}
	if result.EnglishContent.StateLabel == "" {
		result.EnglishContent.StateLabel = "State"
	}
	if result.EnglishContent.ZipCodeLabel == "" {
		result.EnglishContent.ZipCodeLabel = "ZIP Code"
	}
	if result.EnglishContent.AmenitiesLabel == "" {
		result.EnglishContent.AmenitiesLabel = "Amenities & Features"
	}
	if result.EnglishContent.AgentLabel == "" {
		result.EnglishContent.AgentLabel = "Contact Your Agent"
	}
	if result.EnglishContent.PropertyDescriptionLabel == "" {
		result.EnglishContent.PropertyDescriptionLabel = "Property Description"
	}
	if result.EnglishContent.KeyHighlightsLabel == "" {
		result.EnglishContent.KeyHighlightsLabel = "Key Highlights"
	}
	if result.EnglishContent.PropertyGalleryLabel == "" {
		result.EnglishContent.PropertyGalleryLabel = "Property Gallery"
	}
	
	// Arabic fallbacks
	if result.ArabicContent.Title == "" {
		result.ArabicContent.Title = title
	}
	if result.ArabicContent.PriceLabel == "" {
		result.ArabicContent.PriceLabel = "السعر"
	}
	if result.ArabicContent.AddressLabel == "" {
		result.ArabicContent.AddressLabel = "العنوان"
	}
	if result.ArabicContent.CityLabel == "" {
		result.ArabicContent.CityLabel = "المدينة"
	}
	if result.ArabicContent.StateLabel == "" {
		result.ArabicContent.StateLabel = "الولاية"
	}
	if result.ArabicContent.ZipCodeLabel == "" {
		result.ArabicContent.ZipCodeLabel = "الرمز البريدي"
	}
	if result.ArabicContent.AmenitiesLabel == "" {
		result.ArabicContent.AmenitiesLabel = "المرافق والميزات"
	}
	if result.ArabicContent.AgentLabel == "" {
		result.ArabicContent.AgentLabel = "اتصل بوكيلك"
	}
	if result.ArabicContent.PropertyDescriptionLabel == "" {
		result.ArabicContent.PropertyDescriptionLabel = "وصف العقار"
	}
	if result.ArabicContent.KeyHighlightsLabel == "" {
		result.ArabicContent.KeyHighlightsLabel = "المميزات الرئيسية"
	}
	if result.ArabicContent.PropertyGalleryLabel == "" {
		result.ArabicContent.PropertyGalleryLabel = "معرض العقار"
	}

	return &result, nil
}

