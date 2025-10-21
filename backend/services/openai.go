package services

import (
	"context"
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

func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		client: openai.NewClient(apiKey),
	}
}

func (s *OpenAIService) GeneratePropertyContent(title, description, price, currency string, amenities []string) (*AIGeneratedContent, error) {
	ctx := context.Background()

	// Generate English description if not provided or enhance if provided
	englishDesc := description
	if description == "" || len(description) < 50 {
		prompt := fmt.Sprintf(`Generate an engaging and professional property description in English for a real estate listing with the following details:
- Title: %s
- Price: %s %s
- Amenities: %s

The description should be 3-4 paragraphs long, highlight the key features, and appeal to potential buyers. Make it compelling and professional.`, 
			title, price, currency, strings.Join(amenities, ", "))

		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: openai.GPT4,
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
		Model: openai.GPT4,
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

