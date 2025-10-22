package services

import (
	"bytes"
	"fmt"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "io"
	"net/http"
    "os"
	"property-brochure-backend/models"
	"strings"

	"github.com/jung-kurt/gofpdf"
    "golang.org/x/text/encoding/charmap"
    "golang.org/x/text/transform"
)


const (
	// Primary colors
	darkBlueR, darkBlueG, darkBlueB = 31, 78, 121   
	goldR, goldG, goldB             = 212, 175, 55  
	
	// Secondary colors
	lightGrayR, lightGrayG, lightGrayB = 245, 245, 245 
	darkGrayR, darkGrayG, darkGrayB    = 60, 60, 60    
	mediumGrayR, mediumGrayG, mediumGrayB = 120, 120, 120 
	
	// Background colors - warm cream/beige for professional look
	bgCreamR, bgCreamG, bgCreamB = 250, 248, 243
	
	// Page dimensions
	pageWidth  = 210.0
	pageHeight = 297.0
	marginX    = 15.0
	marginY    = 15.0
	contentWidth = pageWidth - (2 * marginX)
)

type PDFService struct{
    arabicFontName string
    hasArabicFont  bool
    brandLogoURL   string
    bodyFontName   string
    hasBodyFont    bool
}

func NewPDFService() *PDFService {
    // Optional branding logo via env var
    logoURL := os.Getenv("BRAND_LOGO_URL")
    return &PDFService{brandLogoURL: logoURL}
}

func (s *PDFService) GenerateBrochure(property *models.Property) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 15) 
    s.setupFonts(pdf)
	
	// Page 1: Cover Page
	s.addCoverPage(pdf, property)
	
	// Page 2: Property Description & Details (English)
	s.addDetailsPageOnly(pdf, property, false)
	
	// Page 3: Investment Opportunity & Gallery
	s.addInvestmentAndGalleryPage(pdf, property, false)
	
	// Page 4: Arabic Description & Agent Contact Info
	s.addArabicAndContactPage(pdf, property)
	
	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateEnglishBrochure creates an English-only brochure
func (s *PDFService) GenerateEnglishBrochure(property *models.Property) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 15)
	s.setupFonts(pdf)
	
	// Page 1: Cover Page
	s.addCoverPage(pdf, property)
	
	// Page 2: Property Description & Details (Description, Highlights, Amenities)
	s.addDetailsPageOnly(pdf, property, false)
	
	// Page 3: Investment Opportunity & Gallery
	s.addInvestmentAndGalleryPage(pdf, property, false)
	
	// Page 4: Agent Contact Info & Thank You
	s.addContactPage(pdf, property)
	
	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate English PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateArabicBrochure creates an Arabic-only brochure with RTL layout
func (s *PDFService) GenerateArabicBrochure(property *models.Property) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 15)
	s.setupFonts(pdf)
	
	// Page 1: Cover Page (Arabic-focused)
	s.addCoverPageArabic(pdf, property)
	
	// Page 2: Arabic Description & Details (Description, Highlights, Amenities)
	s.addDetailsPageOnly(pdf, property, true)
	
	// Page 3: Investment Opportunity & Gallery
	s.addInvestmentAndGalleryPage(pdf, property, true)
	
	// Page 4: Agent Contact Info & Thank You (Arabic labels)
	s.addContactPageWithLanguage(pdf, property, true)
	
	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Arabic PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addCoverPage creates an attractive cover page with main image, title, and price
func (s *PDFService) addCoverPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
	
	// Add cream background to entire page
	s.addPageBackground(pdf)
	
    s.addBrandingIfAvailable(pdf)
	
	// Add decorative corner elements
	s.addDecorativeCorners(pdf)
	
	// Add "Property Brochure" heading at the top
	pdf.SetY(10)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.CellFormat(contentWidth, 8, "Property Brochure", "", 1, "C", false, 0, "")
	
	// Add gold accent bar below heading
	pdf.SetFillColor(goldR, goldG, goldB)
	pdf.Rect(marginX+40, 19, contentWidth-80, 2, "F")
	
	// Add main property image (large, full-width)
	imageHeight := 155.0
	imageStartY := 26.0
	if len(property.ImageURLs) > 0 {
		// Add decorative border around image
		pdf.SetDrawColor(goldR, goldG, goldB)
		pdf.SetLineWidth(1.5)
		pdf.Rect(marginX-1, imageStartY-1, contentWidth+2, imageHeight+2, "D")
		
		// Add image with slight margins
		err := s.addImageFromURL(pdf, property.ImageURLs[0], marginX, imageStartY, contentWidth, imageHeight)
		if err != nil {
			// If image fails, create a placeholder
			pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
			pdf.Rect(marginX, imageStartY, contentWidth, imageHeight, "F")
			pdf.SetFont("Arial", "I", 12)
			pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
			pdf.SetXY(marginX, imageStartY+imageHeight/2)
			pdf.CellFormat(contentWidth, 10, "Image Not Available", "", 0, "C", false, 0, "")
		}
	} else {
		// Placeholder for missing image
		pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
		pdf.Rect(marginX, imageStartY, contentWidth, imageHeight, "F")
		pdf.SetFont("Arial", "I", 12)
		pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
		pdf.SetXY(marginX, imageStartY+imageHeight/2)
		pdf.CellFormat(contentWidth, 10, "No Image Available", "", 0, "C", false, 0, "")
	}
	
	// Property Title (large, bold, dark blue)
	pdf.SetY(186)
	pdf.SetFont("Arial", "B", 26)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	
	// Handle long titles
	titleLines := pdf.SplitLines([]byte(property.Title), contentWidth)
	for _, line := range titleLines {
		pdf.CellFormat(contentWidth, 12, string(line), "", 1, "C", false, 0, "")
	}
	pdf.Ln(3)
	
	// Add a subtle price background box for emphasis
	priceBoxY := pdf.GetY()
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(marginX+35, priceBoxY-2, contentWidth-70, 18, "F")
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Rect(marginX+35, priceBoxY-2, contentWidth-70, 18, "D")
	
	// Price (prominent, gold color)
	pdf.SetY(priceBoxY)
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(goldR, goldG, goldB)
	priceText := s.formatPrice(property.Price, property.Currency)
	pdf.CellFormat(contentWidth, 14, priceText, "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Location (gray, medium size)
	pdf.SetFont("Arial", "", 13)
	pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
	locationText := s.formatLocation(property)
	pdf.MultiCell(contentWidth, 6, locationText, "", "C", false)
	
	// Decorative bottom section with elegant design
	pdf.SetY(268)
	
	// Add decorative diamond shape in center
	centerX := pageWidth / 2
	diamondY := 272.0
	pdf.SetFillColor(goldR, goldG, goldB)
	// Create diamond with lines
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Line(centerX-4, diamondY, centerX, diamondY-3)
	pdf.Line(centerX, diamondY-3, centerX+4, diamondY)
	pdf.Line(centerX+4, diamondY, centerX, diamondY+3)
	pdf.Line(centerX, diamondY+3, centerX-4, diamondY)
	
	// Lines extending from diamond
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX+50, diamondY, centerX-6, diamondY)
	pdf.Line(centerX+6, diamondY, pageWidth-marginX-50, diamondY)
	
	// Add page number
	s.addPageNumber(pdf, 1)
}

// addDetailsPageOnly creates page 2 with only description, highlights, and amenities
func (s *PDFService) addDetailsPageOnly(pdf *gofpdf.Fpdf, property *models.Property, isArabic bool) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	if isArabic {
		s.addArabicDetailsContent(pdf, property, &currentY)
	} else {
		s.addEnglishDetailsContent(pdf, property, &currentY)
	}
	
	// Add decorative bottom diamond element
	s.addBottomDiamondDecoration(pdf)
	
	// Add page number
	s.addPageNumber(pdf, 2)
}

// addEnglishDetailsContent adds English description, highlights, and amenities
func (s *PDFService) addEnglishDetailsContent(pdf *gofpdf.Fpdf, property *models.Property, currentY *float64) {
	// Use localized content if available, fallback to legacy
	var descLabel, highlightsLabel, amenitiesLabel string
	var description string
	var highlights []string
	var amenities []string
	
	if property.EnglishContent.Description != "" {
		// Use new localized content
		descLabel = property.EnglishContent.PropertyDescriptionLabel
		highlightsLabel = property.EnglishContent.KeyHighlightsLabel
		amenitiesLabel = property.EnglishContent.AmenitiesLabel
		description = property.EnglishContent.Description
		highlights = property.EnglishContent.Highlights
		amenities = property.EnglishContent.Amenities
	} else {
		// Fallback to legacy fields
		descLabel = "Property Description"
		highlightsLabel = "Key Highlights"
		amenitiesLabel = "Amenities & Features"
		description = property.AIContent.EnglishDescription
		if description == "" {
			description = property.Description
		}
		highlights = property.AIContent.KeyHighlights
		amenities = property.Amenities
	}
	
	if description == "" {
		description = "No description available."
	}
	
	// Section: Property Description
	*currentY = s.addSectionHeader(pdf, descLabel, *currentY)
	
    if s.hasBodyFont {
        pdf.SetFont(s.bodyFontName, "", 11)
    } else {
        pdf.SetFont("Arial", "", 11)
    }
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, *currentY)
	
	pdf.MultiCell(contentWidth, 5.5, description, "", "L", false)
	*currentY = pdf.GetY() + 8
	
    // Section: Key Highlights
	if len(highlights) > 0 {
		*currentY = s.addSectionHeader(pdf, highlightsLabel, *currentY)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
        for _, raw := range highlights {
            highlight := s.sanitizeBulletText(raw)
            // Draw a gold bullet (filled circle) to avoid Unicode bullet issues
            bulletX := marginX + 5
            bulletY := *currentY + 3.5
            pdf.SetFillColor(goldR, goldG, goldB)
            pdf.Circle(bulletX, bulletY, 1.6, "F")

            // Highlight text
            pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
            pdf.SetFont("Arial", "", 11)
            pdf.SetXY(marginX+12, *currentY)
            pdf.MultiCell(contentWidth-12, 6, highlight, "", "L", false)
            *currentY = pdf.GetY() + 1
        }
		*currentY += 6
	}
	
	// Section: Amenities
	if len(amenities) > 0 {
		// Check if we need space on page
		if *currentY > 220 {
			// Skip to make room - we won't add a new page, just adjust spacing
			*currentY = 220
		}
		
		*currentY = s.addSectionHeader(pdf, amenitiesLabel, *currentY)
		
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
        // Display amenities in a 2-column grid with checkmarks
		colWidth := (contentWidth - 10) / 2
		amenityHeight := 7.0
		
		for i, amenity := range amenities {
			col := i % 2
			xPos := marginX + float64(col)*(colWidth+10)
			
			pdf.SetXY(xPos, *currentY)
			
            // Draw a green check mark using vector lines (avoids Unicode glyph issues)
            pdf.SetDrawColor(46, 125, 50)
            pdf.SetLineWidth(0.8)
            startX := xPos
            startY := *currentY + amenityHeight/2
            pdf.Line(startX, startY, startX+2.0, startY+2.0)
            pdf.Line(startX+2.0, startY+2.0, startX+6.0, startY-1.0)
			
            // Amenity text
            pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
            if s.hasBodyFont {
                pdf.SetFont(s.bodyFontName, "", 10)
            } else {
                pdf.SetFont("Arial", "", 10)
            }
            pdf.SetX(xPos + 9)
			pdf.CellFormat(colWidth-7, amenityHeight, amenity, "", 0, "", false, 0, "")
			
			// Move to next row after 2 columns
			if col == 1 {
				*currentY += amenityHeight
			}
		}
		
		// Handle odd number of amenities
		if len(amenities)%2 == 1 {
			*currentY += amenityHeight
		}
	}
}

// addArabicDetailsContent adds Arabic description, highlights, and amenities
func (s *PDFService) addArabicDetailsContent(pdf *gofpdf.Fpdf, property *models.Property, currentY *float64) {
	// Use localized content if available, fallback to legacy
	var descLabel, highlightsLabel, amenitiesLabel string
	var description string
	var highlights []string
	var amenities []string
	
	if property.ArabicContent.Description != "" {
		// Use new localized content
		descLabel = property.ArabicContent.PropertyDescriptionLabel
		highlightsLabel = property.ArabicContent.KeyHighlightsLabel
		amenitiesLabel = property.ArabicContent.AmenitiesLabel
		description = property.ArabicContent.Description
		highlights = property.ArabicContent.Highlights
		amenities = property.ArabicContent.Amenities
	} else {
		// Fallback to legacy fields
		descLabel = "وصف العقار"
		highlightsLabel = "المميزات الرئيسية"
		amenitiesLabel = "المرافق والميزات"
		description = property.AIContent.ArabicDescription
		highlights = []string{}
		amenities = property.Amenities
	}
	
	if description == "" {
		description = "لا يوجد وصف متاح"
	}
	
	// Section: Arabic Description
	if s.hasArabicFont {
		*currentY = s.addSectionHeaderAligned(pdf, descLabel, *currentY, s.arabicFontName, "R")
	} else {
		*currentY = s.addSectionHeader(pdf, descLabel, *currentY)
	}
	
	// Use Arabic font if available
	if s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 12)
	} else {
		pdf.SetFont("Arial", "", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, *currentY)
	
	// Right-aligned for Arabic text
	description = s.fixMojibakeLatin1ToUTF8(description)
	pdf.MultiCell(contentWidth, 6, description, "", "R", false)
	*currentY = pdf.GetY() + 8
	
	// Section: Key Highlights (Arabic)
	if len(highlights) > 0 {
		if s.hasArabicFont {
			*currentY = s.addSectionHeaderAligned(pdf, highlightsLabel, *currentY, s.arabicFontName, "R")
		} else {
			*currentY = s.addSectionHeader(pdf, highlightsLabel, *currentY)
		}
		
		if s.hasArabicFont {
			pdf.SetFont(s.arabicFontName, "", 11)
		} else {
			pdf.SetFont("Arial", "", 11)
		}
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
		for _, raw := range highlights {
			highlight := s.sanitizeBulletText(raw)
			highlight = s.fixMojibakeLatin1ToUTF8(highlight)
			
			// Draw a gold bullet (filled circle)
			bulletX := pageWidth - marginX - 5 // Right side for RTL
			bulletY := *currentY + 3.5
			pdf.SetFillColor(goldR, goldG, goldB)
			pdf.Circle(bulletX, bulletY, 1.6, "F")
			
			// Highlight text (right-aligned)
			pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
			if s.hasArabicFont {
				pdf.SetFont(s.arabicFontName, "", 11)
			} else {
				pdf.SetFont("Arial", "", 11)
			}
			pdf.SetXY(marginX, *currentY)
			pdf.MultiCell(contentWidth-12, 6, highlight, "", "R", false)
			*currentY = pdf.GetY() + 1
		}
		*currentY += 6
	}
	
	// Section: Amenities (if available)
	if len(amenities) > 0 {
		// Check if we need space on page
		if *currentY > 220 {
			*currentY = 220
		}
		
		if s.hasArabicFont {
			*currentY = s.addSectionHeaderAligned(pdf, amenitiesLabel, *currentY, s.arabicFontName, "R")
		} else {
			*currentY = s.addSectionHeader(pdf, amenitiesLabel, *currentY)
		}
		
		if s.hasArabicFont {
			pdf.SetFont(s.arabicFontName, "", 10)
		} else {
			pdf.SetFont("Arial", "", 10)
		}
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
		// Display amenities in a 2-column grid with checkmarks
		colWidth := (contentWidth - 10) / 2
		amenityHeight := 7.0
		
		for i, amenity := range amenities {
			col := i % 2
			xPos := marginX + float64(col)*(colWidth+10)
			
			pdf.SetXY(xPos, *currentY)
			
			// Draw a green check mark using vector lines
			pdf.SetDrawColor(46, 125, 50)
			pdf.SetLineWidth(0.8)
			startX := xPos
			startY := *currentY + amenityHeight/2
			pdf.Line(startX, startY, startX+2.0, startY+2.0)
			pdf.Line(startX+2.0, startY+2.0, startX+6.0, startY-1.0)
			
			// Amenity text (apply mojibake fix for Arabic)
			amenity = s.fixMojibakeLatin1ToUTF8(amenity)
			pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
			if s.hasArabicFont {
				pdf.SetFont(s.arabicFontName, "", 10)
			} else {
				pdf.SetFont("Arial", "", 10)
			}
			pdf.SetX(xPos + 9)
			pdf.CellFormat(colWidth-7, amenityHeight, amenity, "", 0, "", false, 0, "")
			
			// Move to next row after 2 columns
			if col == 1 {
				*currentY += amenityHeight
			}
		}
		
		// Handle odd number of amenities
		if len(amenities)%2 == 1 {
			*currentY += amenityHeight
		}
	}
}

// addInvestmentAndGalleryPage creates page 3 with investment opportunity and property gallery
func (s *PDFService) addInvestmentAndGalleryPage(pdf *gofpdf.Fpdf, property *models.Property, isArabic bool) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	// Section: Investment Opportunity
	var additionalTitle, additionalContent string
	if isArabic {
		if property.ArabicContent.AdditionalSectionTitle != "" {
			additionalTitle = property.ArabicContent.AdditionalSectionTitle
			additionalContent = property.ArabicContent.AdditionalSectionContent
		} else {
			additionalTitle = "فرصة استثمارية"
			additionalContent = "يمثل هذا العقار فرصة استثمارية ممتازة."
		}
		additionalTitle = s.fixMojibakeLatin1ToUTF8(additionalTitle)
		additionalContent = s.fixMojibakeLatin1ToUTF8(additionalContent)
	} else {
		if property.EnglishContent.AdditionalSectionTitle != "" {
			additionalTitle = property.EnglishContent.AdditionalSectionTitle
			additionalContent = property.EnglishContent.AdditionalSectionContent
		} else {
			additionalTitle = "Investment Opportunity"
			additionalContent = "This property represents an excellent investment opportunity."
		}
	}
	
	if additionalContent != "" {
		if isArabic && s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, additionalTitle, currentY, s.arabicFontName, "R")
			pdf.SetFont(s.arabicFontName, "", 11)
		} else {
			currentY = s.addSectionHeaderWithIcon(pdf, additionalTitle, currentY, "investment")
			if s.hasBodyFont {
				pdf.SetFont(s.bodyFontName, "", 11)
			} else {
				pdf.SetFont("Arial", "", 11)
			}
		}
		
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		pdf.SetXY(marginX, currentY)
		align := "L"
		if isArabic {
			align = "R"
		}
		pdf.MultiCell(contentWidth, 5.5, additionalContent, "", align, false)
		currentY = pdf.GetY() + 12
	}
	
	// Add Property Gallery (if images available)
	if len(property.ImageURLs) > 1 {
		galleryLabel := "Property Gallery"
		if isArabic {
			if property.ArabicContent.PropertyGalleryLabel != "" {
				galleryLabel = property.ArabicContent.PropertyGalleryLabel
			} else {
				galleryLabel = "معرض العقار"
			}
			galleryLabel = s.fixMojibakeLatin1ToUTF8(galleryLabel)
		} else {
			if property.EnglishContent.PropertyGalleryLabel != "" {
				galleryLabel = property.EnglishContent.PropertyGalleryLabel
			}
		}
		
		if isArabic && s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, galleryLabel, currentY, s.arabicFontName, "R")
		} else {
			currentY = s.addSectionHeaderWithIcon(pdf, galleryLabel, currentY, "gallery")
		}
		currentY += 3
		
		// Display up to 4 additional images in a compact 2x2 grid
		imgWidth := (contentWidth - 8) / 2
		imgHeight := imgWidth * 0.65
		spacing := 8.0
		
		imageCount := 0
		maxImages := 4
		
		for i := 1; i < len(property.ImageURLs) && imageCount < maxImages; i++ {
			row := imageCount / 2
			col := imageCount % 2
			
			xPos := marginX + float64(col)*(imgWidth+spacing)
			yPos := currentY + float64(row)*(imgHeight+spacing)
			
			// Check if we're running out of space
			if yPos+imgHeight > pageHeight-35 {
				break
			}
			
			// Add shadow effect
			pdf.SetFillColor(180, 180, 180)
			pdf.Rect(xPos+1.5, yPos+1.5, imgWidth, imgHeight, "F")
			
			// Add white background
			pdf.SetFillColor(255, 255, 255)
			pdf.Rect(xPos, yPos, imgWidth, imgHeight, "F")
			
			// Add gold border/frame effect
			pdf.SetDrawColor(goldR, goldG, goldB)
			pdf.SetLineWidth(0.6)
			pdf.Rect(xPos, yPos, imgWidth, imgHeight, "D")
			
			err := s.addImageFromURL(pdf, property.ImageURLs[i], xPos+2, yPos+2, imgWidth-4, imgHeight-4)
			if err != nil {
				// Placeholder for failed images
				pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
				pdf.Rect(xPos+2, yPos+2, imgWidth-4, imgHeight-4, "F")
			}
			
			imageCount++
		}
	}
	
	// Add decorative bottom diamond element
	s.addBottomDiamondDecoration(pdf)
	
	// Add page number
	s.addPageNumber(pdf, 3)
}

// addGalleryPage creates an image gallery for additional property photos
func (s *PDFService) addGalleryPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	// Use localized label if available
	galleryLabel := "Property Gallery"
	if property.EnglishContent.PropertyGalleryLabel != "" {
		galleryLabel = property.EnglishContent.PropertyGalleryLabel
	}
	
	// Section header
	currentY = s.addSectionHeader(pdf, galleryLabel, currentY)
	currentY += 5
	
	// Display up to 4 additional images in a 2x2 grid
	imgWidth := (contentWidth - 10) / 2
	imgHeight := imgWidth * 0.75 // 4:3 aspect ratio
		spacing := 10.0

	imageCount := 0
	maxImages := 4
	
	for i := 1; i < len(property.ImageURLs) && imageCount < maxImages; i++ {
		row := imageCount / 2
		col := imageCount % 2
		
		xPos := marginX + float64(col)*(imgWidth+spacing)
		yPos := currentY + float64(row)*(imgHeight+spacing)
		
		// Add shadow effect
		pdf.SetFillColor(180, 180, 180)
		pdf.Rect(xPos+2, yPos+2, imgWidth, imgHeight, "F")
		
		// Add white background
		pdf.SetFillColor(255, 255, 255)
		pdf.Rect(xPos, yPos, imgWidth, imgHeight, "F")
		
		// Add gold border/frame effect
		pdf.SetDrawColor(goldR, goldG, goldB)
		pdf.SetLineWidth(0.8)
		pdf.Rect(xPos, yPos, imgWidth, imgHeight, "D")
		
		err := s.addImageFromURL(pdf, property.ImageURLs[i], xPos+2, yPos+2, imgWidth-4, imgHeight-4)
		if err != nil {
			// Placeholder for failed images
			pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
			pdf.Rect(xPos+2, yPos+2, imgWidth-4, imgHeight-4, "F")
		}
		
		imageCount++
	}
	
	// Add page number
	s.addPageNumber(pdf, 3)
}

// addArabicAndContactPage creates the Arabic description and agent contact page
func (s *PDFService) addArabicAndContactPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
    // Section: Arabic Description (use Arabic font and right alignment if available)
    headerTextAr := "وصف العقار"
    if s.hasArabicFont {
        currentY = s.addSectionHeaderAligned(pdf, headerTextAr, currentY, s.arabicFontName, "R")
    } else {
        currentY = s.addSectionHeader(pdf, "Arabic Description", currentY)
    }
	
    // Use Arabic font if available
    if s.hasArabicFont {
        pdf.SetFont(s.arabicFontName, "", 12)
    } else {
        if s.hasBodyFont {
            pdf.SetFont(s.bodyFontName, "", 11)
        } else {
            pdf.SetFont("Arial", "", 11)
        }
    }
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, currentY)
	
    arabicDesc := property.AIContent.ArabicDescription
	if arabicDesc == "" {
		arabicDesc = "لا يوجد وصف متاح"
	}
	
    // Right-aligned for Arabic text (ensure UTF-8 font and R align). Apply shaping if font is present.
    arabicDesc = s.fixMojibakeLatin1ToUTF8(arabicDesc)
    pdf.MultiCell(contentWidth, 6, arabicDesc, "", "R", false)
	currentY = pdf.GetY() + 15
	
	// Agent Contact Card - positioned at top section instead of bottom
	currentY = s.addAgentContactCardTop(pdf, property, currentY, false)
	
	// Add spacing
	currentY += 15
	
	// Add thank you message
	s.addThankYouMessage(pdf, property, currentY, false)
	
	// Add decorative bottom diamond element
	s.addBottomDiamondDecoration(pdf)
	
	// Add page number (now page 4 with restructuring)
	s.addPageNumber(pdf, 4)
}

// addAgentContactCard creates a professional contact card for the agent (English)
func (s *PDFService) addAgentContactCard(pdf *gofpdf.Fpdf, property *models.Property, startY float64) {
	s.addAgentContactCardLocalized(pdf, property, startY, false)
}

// addAgentContactCardLocalized creates a professional contact card with optional Arabic labels
func (s *PDFService) addAgentContactCardLocalized(pdf *gofpdf.Fpdf, property *models.Property, startY float64, useArabic bool) {
	cardHeight := 55.0
	cardY := pageHeight - marginY - cardHeight - 20
	
	// Background card
	pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
	pdf.Rect(marginX, cardY, contentWidth, cardHeight, "F")
	
	// Gold accent border
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Rect(marginX, cardY, contentWidth, cardHeight, "D")
	
	// Determine labels based on language
	var agentLabel, nameLabel, emailLabel, phoneLabel string
	var align string
	
	if useArabic && property.ArabicContent.AgentLabel != "" {
		agentLabel = property.ArabicContent.AgentLabel
		nameLabel = "الاسم:"
		emailLabel = "البريد الإلكتروني:"
		phoneLabel = "الهاتف:"
		align = "R"
	} else if !useArabic && property.EnglishContent.AgentLabel != "" {
		agentLabel = property.EnglishContent.AgentLabel
		nameLabel = "Name:"
		emailLabel = "Email:"
		phoneLabel = "Phone:"
		align = "C"
	} else {
		// Fallback to English
		agentLabel = "Contact Your Agent"
		nameLabel = "Name:"
		emailLabel = "Email:"
		phoneLabel = "Phone:"
		align = "C"
	}
	
	// "Contact Agent" header
	pdf.SetXY(marginX+5, cardY+5)
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 14)
	} else {
		pdf.SetFont("Arial", "B", 14)
	}
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	agentLabel = s.fixMojibakeLatin1ToUTF8(agentLabel)
	pdf.CellFormat(contentWidth-10, 8, agentLabel, "", 1, align, false, 0, "")
	
	// Divider line
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.3)
	pdf.Line(marginX+30, cardY+13, pageWidth-marginX-30, cardY+13)
	
	// Agent info
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, cardY+18)
	nameLabel = s.fixMojibakeLatin1ToUTF8(nameLabel)
	pdf.CellFormat(50, 6, nameLabel, "", 0, "", false, 0, "")
	
	if s.hasBodyFont && !useArabic {
		pdf.SetFont(s.bodyFontName, "", 11)
	} else if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "", 11)
	}
	pdf.CellFormat(0, 6, property.AgentInfo.Name, "", 0, "", false, 0, "")
	
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetXY(marginX+10, cardY+28)
	emailLabel = s.fixMojibakeLatin1ToUTF8(emailLabel)
	pdf.CellFormat(50, 6, emailLabel, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.CellFormat(0, 6, property.AgentInfo.Email, "", 0, "", false, 0, "")
	
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, cardY+38)
	phoneLabel = s.fixMojibakeLatin1ToUTF8(phoneLabel)
	pdf.CellFormat(50, 6, phoneLabel, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(goldR, goldG, goldB)
	pdf.CellFormat(0, 6, property.AgentInfo.Phone, "", 0, "", false, 0, "")
}

// addSectionHeader creates a styled section header
func (s *PDFService) addSectionHeader(pdf *gofpdf.Fpdf, title string, y float64) float64 {
	// Background bar
	pdf.SetFillColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.Rect(marginX, y, contentWidth, 10, "F")
	
	// Title text
	pdf.SetXY(marginX+5, y+1.5)
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(255, 255, 255) // White text
	pdf.CellFormat(contentWidth-10, 7, title, "", 0, "L", false, 0, "")
	
	// Gold accent line
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Line(marginX, y+10, pageWidth-marginX, y+10)
	
	return y + 15
}

// addSectionHeaderWithIcon creates an enhanced section header with decorative elements
func (s *PDFService) addSectionHeaderWithIcon(pdf *gofpdf.Fpdf, title string, y float64, iconType string) float64 {
	// Gradient effect using two rectangles
	pdf.SetFillColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.Rect(marginX, y, contentWidth, 10, "F")
	
	// Add decorative left accent bar
	pdf.SetFillColor(goldR, goldG, goldB)
	pdf.Rect(marginX, y, 3, 10, "F")
	
	// Add decorative right corner
	pdf.SetFillColor(goldR-20, goldG-20, goldB-20)
	pdf.Rect(pageWidth-marginX-3, y, 3, 10, "F")
	
	// Icon/bullet point
	iconX := marginX + 8
	iconY := y + 5
	pdf.SetFillColor(goldR, goldG, goldB)
	pdf.Circle(iconX, iconY, 2, "F")
	
	// Title text
	pdf.SetXY(marginX+14, y+1.5)
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(255, 255, 255) // White text
	pdf.CellFormat(contentWidth-20, 7, title, "", 0, "L", false, 0, "")
	
	// Gold accent line with fade effect
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(1.0)
	pdf.Line(marginX, y+10, pageWidth-marginX, y+10)
	
	return y + 15
}

// addSectionHeaderAligned is like addSectionHeader but allows custom font and alignment
func (s *PDFService) addSectionHeaderAligned(pdf *gofpdf.Fpdf, title string, y float64, fontName string, align string) float64 {
    if align != "R" {
        align = "L"
    }
    // Background bar
    pdf.SetFillColor(darkBlueR, darkBlueG, darkBlueB)
    pdf.Rect(marginX, y, contentWidth, 10, "F")

    // Title text with custom font if provided
    pdf.SetTextColor(255, 255, 255)
    if fontName != "" {
        pdf.SetFont(fontName, "", 13)
    } else {
        pdf.SetFont("Arial", "B", 13)
    }

    // Position and alignment
    pdf.SetXY(marginX+5, y+1.5)
    pdf.CellFormat(contentWidth-10, 7, title, "", 0, align, false, 0, "")

    // Gold accent line
    pdf.SetDrawColor(goldR, goldG, goldB)
    pdf.SetLineWidth(0.8)
    pdf.Line(marginX, y+10, pageWidth-marginX, y+10)

    return y + 15
}

// addPageNumber adds page number at the bottom of the page
func (s *PDFService) addPageNumber(pdf *gofpdf.Fpdf, pageNum int) {
	pdf.SetY(-10)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
	pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pageNum), "", 0, "C", false, 0, "")
}

// setupFonts attempts to load optional Unicode fonts for better internationalization
func (s *PDFService) setupFonts(pdf *gofpdf.Fpdf) {
    // Force override: Use hardcoded paths from project fonts folder
    fontPath := "fonts/NotoNaskhArabic-Regular.ttf"
    
    fmt.Println("[PDF DEBUG] Using Arabic font path:", fontPath)
    
    if _, err := os.Stat(fontPath); err == nil {
        pdf.AddUTF8Font("ArabicFont", "", fontPath)
        s.arabicFontName = "ArabicFont"
        s.hasArabicFont = true
        fmt.Println("[PDF] Loaded Arabic UTF-8 font:", fontPath)
    } else {
        fmt.Println("[PDF] ARABIC_TTF_PATH not found:", fontPath, "err:", err)
    }

    // Force override: Use hardcoded paths from project fonts folder
    bodyPath := "fonts/Roboto-Regular.ttf"
    fmt.Println("[PDF DEBUG] Using body font path:", bodyPath)
    
    if _, err := os.Stat(bodyPath); err == nil {
        pdf.AddUTF8Font("BodyFont", "", bodyPath)
        s.bodyFontName = "BodyFont"
        s.hasBodyFont = true
        fmt.Println("[PDF] Loaded Body UTF-8 font:", bodyPath)
    } else {
        fmt.Println("[PDF] BODY_TTF_PATH not found:", bodyPath, "err:", err)
    }

    // Fallback: if body font not set but Arabic font exists, use Arabic font for body too
    if !s.hasBodyFont && s.hasArabicFont {
        s.bodyFontName = s.arabicFontName
        s.hasBodyFont = true
        fmt.Println("[PDF] Using Arabic font as body font fallback.")
    }
}

// addBrandingIfAvailable draws a small logo in the top-right corner if BRAND_LOGO_URL is set
func (s *PDFService) addBrandingIfAvailable(pdf *gofpdf.Fpdf) {
    if s.brandLogoURL == "" {
        return
    }
    // Reserve a small square area for the logo
    boxW, boxH := 18.0, 18.0
    x := pageWidth - marginX - boxW
    y := 6.0
    _ = s.addImageFromURL(pdf, s.brandLogoURL, x, y, boxW, boxH)
}

// formatPrice formats the price with currency symbol
func (s *PDFService) formatPrice(price float64, currency string) string {
	if currency == "" {
		currency = "USD"
	}
	
	// Format with thousand separators
	priceStr := fmt.Sprintf("%.0f", price)
	
	// Add thousand separators
	if len(priceStr) > 3 {
		result := ""
		for i, digit := range priceStr {
			if i > 0 && (len(priceStr)-i)%3 == 0 {
				result += ","
			}
			result += string(digit)
		}
		priceStr = result
	}
	
	return fmt.Sprintf("%s %s", currency, priceStr)
}

// formatLocation creates a formatted location string
func (s *PDFService) formatLocation(property *models.Property) string {
	parts := []string{}
	
	if property.Address != "" {
		parts = append(parts, property.Address)
	}
	if property.City != "" {
		parts = append(parts, property.City)
	}
	if property.State != "" {
		parts = append(parts, property.State)
	}
	if property.ZipCode != "" {
		parts = append(parts, property.ZipCode)
	}
	
	if len(parts) == 0 {
		return "Location not specified"
	}
	
	return strings.Join(parts, ", ")
}

// sanitizeBulletText removes any leading bullet/arrow characters that might be included by AI
func (s *PDFService) sanitizeBulletText(text string) string {
    trimmed := strings.TrimSpace(text)
    // Common bad prefixes: "•", "-", "--", "*", "·", "—", "->", "=>", "â€¢" (mojibake)
    prefixes := []string{"â€¢", "•", "->", "=>", "—", "·", "--", "-", "*"}
    for _, p := range prefixes {
        if strings.HasPrefix(trimmed, p+" ") {
            trimmed = strings.TrimSpace(trimmed[len(p)+1:])
            break
        } else if strings.HasPrefix(trimmed, p) {
            trimmed = strings.TrimSpace(trimmed[len(p):])
            break
        }
    }
    return trimmed
}

// fixMojibakeLatin1ToUTF8 attempts to convert text that was UTF-8 but decoded as Latin-1
// This helps when inputs show sequences like "Ã˜" instead of proper Arabic letters.
func (s *PDFService) fixMojibakeLatin1ToUTF8(text string) string {
    // If text already contains Arabic codepoints, return as-is
    for _, r := range text {
        if r >= 0x0600 && r <= 0x06FF {
            return text
        }
    }
    // Heuristic: if it contains 'Ã' (common mojibake indicator), try Latin-1 decode
    if !strings.ContainsRune(text, 'Ã') {
        return text
    }
    reader := transform.NewReader(strings.NewReader(text), charmap.ISO8859_1.NewDecoder())
    decoded, err := io.ReadAll(reader)
    if err != nil {
        return text
    }
    return string(decoded)
}

// addPageBackground adds a cream-colored background to the entire page
func (s *PDFService) addPageBackground(pdf *gofpdf.Fpdf) {
	pdf.SetFillColor(bgCreamR, bgCreamG, bgCreamB)
	pdf.Rect(0, 0, pageWidth, pageHeight, "F")
}

// addDecorativeCorners adds decorative corner elements to the page
func (s *PDFService) addDecorativeCorners(pdf *gofpdf.Fpdf) {
	// Top-left corner
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.5)
	pdf.Line(5, 5, 15, 5)
	pdf.Line(5, 5, 5, 15)
	
	// Top-right corner
	pdf.Line(pageWidth-15, 5, pageWidth-5, 5)
	pdf.Line(pageWidth-5, 5, pageWidth-5, 15)
	
	// Bottom-left corner
	pdf.Line(5, pageHeight-15, 5, pageHeight-5)
	pdf.Line(5, pageHeight-5, 15, pageHeight-5)
	
	// Bottom-right corner
	pdf.Line(pageWidth-15, pageHeight-5, pageWidth-5, pageHeight-5)
	pdf.Line(pageWidth-5, pageHeight-15, pageWidth-5, pageHeight-5)
}

// addBottomDiamondDecoration adds the elegant diamond with lines decoration at the bottom of the page
func (s *PDFService) addBottomDiamondDecoration(pdf *gofpdf.Fpdf) {
	// Position near bottom but above page number
	pdf.SetY(268)
	
	// Add decorative diamond shape in center
	centerX := pageWidth / 2
	diamondY := 272.0
	pdf.SetFillColor(goldR, goldG, goldB)
	
	// Create diamond with lines
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Line(centerX-4, diamondY, centerX, diamondY-3)
	pdf.Line(centerX, diamondY-3, centerX+4, diamondY)
	pdf.Line(centerX+4, diamondY, centerX, diamondY+3)
	pdf.Line(centerX, diamondY+3, centerX-4, diamondY)
	
	// Lines extending from diamond
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX+50, diamondY, centerX-6, diamondY)
	pdf.Line(centerX+6, diamondY, pageWidth-marginX-50, diamondY)
}

// addAgentContactCardTop creates a professional contact card at the top of the page and returns the Y position after the card
func (s *PDFService) addAgentContactCardTop(pdf *gofpdf.Fpdf, property *models.Property, startY float64, useArabic bool) float64 {
	cardHeight := 55.0
	
	// Background card with shadow effect
	pdf.SetFillColor(200, 200, 200)
	pdf.Rect(marginX+2, startY+2, contentWidth, cardHeight, "F")
	
	// Main card background
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(marginX, startY, contentWidth, cardHeight, "F")
	
	// Gold accent border
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Rect(marginX, startY, contentWidth, cardHeight, "D")
	
	// Determine labels based on language
	var agentLabel, nameLabel, emailLabel, phoneLabel string
	var align string
	
	if useArabic && property.ArabicContent.AgentLabel != "" {
		agentLabel = property.ArabicContent.AgentLabel
		nameLabel = "الاسم:"
		emailLabel = "البريد الإلكتروني:"
		phoneLabel = "الهاتف:"
		align = "R"
	} else if !useArabic && property.EnglishContent.AgentLabel != "" {
		agentLabel = property.EnglishContent.AgentLabel
		nameLabel = "Name:"
		emailLabel = "Email:"
		phoneLabel = "Phone:"
		align = "C"
	} else {
		// Fallback to English
		agentLabel = "Contact Your Agent"
		nameLabel = "Name:"
		emailLabel = "Email:"
		phoneLabel = "Phone:"
		align = "C"
	}
	
	// "Contact Agent" header
	pdf.SetXY(marginX+5, startY+5)
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 14)
	} else {
		pdf.SetFont("Arial", "B", 14)
	}
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	agentLabel = s.fixMojibakeLatin1ToUTF8(agentLabel)
	pdf.CellFormat(contentWidth-10, 8, agentLabel, "", 1, align, false, 0, "")
	
	// Divider line
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.3)
	pdf.Line(marginX+30, startY+13, pageWidth-marginX-30, startY+13)
	
	// Agent info
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, startY+18)
	nameLabel = s.fixMojibakeLatin1ToUTF8(nameLabel)
	pdf.CellFormat(50, 6, nameLabel, "", 0, "", false, 0, "")
	
	if s.hasBodyFont && !useArabic {
		pdf.SetFont(s.bodyFontName, "", 11)
	} else if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "", 11)
	}
	pdf.CellFormat(0, 6, property.AgentInfo.Name, "", 0, "", false, 0, "")
	
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetXY(marginX+10, startY+28)
	emailLabel = s.fixMojibakeLatin1ToUTF8(emailLabel)
	pdf.CellFormat(50, 6, emailLabel, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.CellFormat(0, 6, property.AgentInfo.Email, "", 0, "", false, 0, "")
	
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "B", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, startY+38)
	phoneLabel = s.fixMojibakeLatin1ToUTF8(phoneLabel)
	pdf.CellFormat(50, 6, phoneLabel, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(goldR, goldG, goldB)
	pdf.CellFormat(0, 6, property.AgentInfo.Phone, "", 0, "", false, 0, "")
	
	return startY + cardHeight
}

// addThankYouMessage adds a thank you message section below the agent card
func (s *PDFService) addThankYouMessage(pdf *gofpdf.Fpdf, property *models.Property, startY float64, useArabic bool) {
	var thankYouMsg string
	var align string
	
	if useArabic && property.ArabicContent.ThankYouMessage != "" {
		thankYouMsg = property.ArabicContent.ThankYouMessage
		align = "R"
	} else if !useArabic && property.EnglishContent.ThankYouMessage != "" {
		thankYouMsg = property.EnglishContent.ThankYouMessage
		align = "L"
	} else {
		// Fallback
		if useArabic {
			thankYouMsg = "نشكركم على اهتمامكم بهذا العقار الاستثنائي. نحن نقدر اهتمامكم ويسعدنا تزويدكم بمعلومات إضافية أو ترتيب موعد للمعاينة في الوقت المناسب لكم."
			align = "R"
		} else {
			thankYouMsg = "Thank you for considering this exceptional property. We appreciate your interest and would be delighted to provide you with additional information or arrange a viewing at your convenience."
			align = "L"
		}
	}
	
	// Add simple decorative line (thin gold line only)
	pdf.SetY(startY)
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX+contentWidth/2-30, startY, marginX+contentWidth/2+30, startY)
	
	startY += 10
	
	// Add thank you message
	if useArabic && s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 12)
	} else if s.hasBodyFont {
		pdf.SetFont(s.bodyFontName, "", 11)
	} else {
		pdf.SetFont("Arial", "", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, startY)
	
	thankYouMsg = s.fixMojibakeLatin1ToUTF8(thankYouMsg)
	pdf.MultiCell(contentWidth, 6, thankYouMsg, "", align, false)
	
}


func (s *PDFService) addImageFromURL(pdf *gofpdf.Fpdf, url string, x, y, w, h float64) error {
	// Download image
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

    // Read the body into memory so we can decode dimensions and also register with gofpdf
    var imgBuf bytes.Buffer
    if _, err := io.Copy(&imgBuf, resp.Body); err != nil {
        return err
    }

	// Determine image type from content type
	imageType := "jpg"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "png") {
		imageType = "png"
	} else if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		imageType = "jpg"
	}

    // Decode to get intrinsic dimensions
    imgReader := bytes.NewReader(imgBuf.Bytes())
    decoded, _, err := image.Decode(imgReader)
    if err != nil {
        // If decode fails, still try to place the image without aspect fit
        imgReader = bytes.NewReader(imgBuf.Bytes())
    } else {
        // Calculate aspect-fit size
        imgW := float64(decoded.Bounds().Dx())
        imgH := float64(decoded.Bounds().Dy())
        if imgW > 0 && imgH > 0 {
            scale := w / imgW
            if imgH*scale > h {
                scale = h / imgH
            }
            drawW := imgW * scale
            drawH := imgH * scale
            // center within the box
            x = x + (w-drawW)/2
            y = y + (h-drawH)/2
            w = drawW
            h = drawH
        }
        // reset reader for registration
        imgReader = bytes.NewReader(imgBuf.Bytes())
    }

	// Create unique name for this image
	urlSuffix := url
	if len(url) > 20 {
		urlSuffix = url[len(url)-20:]
	}
	uniqueName := fmt.Sprintf("img_%s_%.0f_%.0f", urlSuffix, x, y)

	// Register and add image to PDF using ImageOptions
	opts := gofpdf.ImageOptions{
		ImageType:             imageType,
		ReadDpi:               false,
		AllowNegativePosition: false,
	}
    pdf.RegisterImageOptionsReader(uniqueName, opts, imgReader)
	pdf.ImageOptions(uniqueName, x, y, w, h, false, opts, 0, "")

	return nil
}

// addContactPage creates a standalone contact page (without Arabic description)
func (s *PDFService) addContactPage(pdf *gofpdf.Fpdf, property *models.Property) {
	s.addContactPageWithLanguage(pdf, property, false)
}

// addContactPageWithLanguage creates a standalone contact page with language support
func (s *PDFService) addContactPageWithLanguage(pdf *gofpdf.Fpdf, property *models.Property, useArabic bool) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
	s.addBrandingIfAvailable(pdf)
	
	currentY := marginY + 10.0
	
	// Agent Contact Card at the top
	currentY = s.addAgentContactCardTop(pdf, property, currentY, useArabic)
	
	// Add spacing
	currentY += 15
	
	// Add thank you message below agent card
	s.addThankYouMessage(pdf, property, currentY, useArabic)
	
	// Add decorative bottom diamond element
	s.addBottomDiamondDecoration(pdf)
	
	// Add page number (now page 4 with restructuring)
	s.addPageNumber(pdf, 4)
}

// addCoverPageArabic creates an Arabic-focused cover page
func (s *PDFService) addCoverPageArabic(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
	s.addBrandingIfAvailable(pdf)
	
	// Add decorative corner elements
	s.addDecorativeCorners(pdf)
	
	// Add "Property Brochure" heading in Arabic
	pdf.SetY(10)
	if s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 16)
	} else {
		pdf.SetFont("Arial", "B", 16)
	}
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	brochureLabel := "كتيب العقار"
	brochureLabel = s.fixMojibakeLatin1ToUTF8(brochureLabel)
	pdf.CellFormat(contentWidth, 8, brochureLabel, "", 1, "C", false, 0, "")
	
	// Add gold accent bar below heading
	pdf.SetFillColor(goldR, goldG, goldB)
	pdf.Rect(marginX+40, 19, contentWidth-80, 2, "F")
	
	// Add main property image (large, full-width)
	imageHeight := 155.0
	imageStartY := 26.0
	if len(property.ImageURLs) > 0 {
		// Add decorative border around image
		pdf.SetDrawColor(goldR, goldG, goldB)
		pdf.SetLineWidth(1.5)
		pdf.Rect(marginX-1, imageStartY-1, contentWidth+2, imageHeight+2, "D")
		
		err := s.addImageFromURL(pdf, property.ImageURLs[0], marginX, imageStartY, contentWidth, imageHeight)
		if err != nil {
			pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
			pdf.Rect(marginX, imageStartY, contentWidth, imageHeight, "F")
			pdf.SetFont("Arial", "I", 12)
			pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
			pdf.SetXY(marginX, imageStartY+imageHeight/2)
			pdf.CellFormat(contentWidth, 10, "Image Not Available", "", 0, "C", false, 0, "")
		}
	} else {
		pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
		pdf.Rect(marginX, imageStartY, contentWidth, imageHeight, "F")
		pdf.SetFont("Arial", "I", 12)
		pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
		pdf.SetXY(marginX, imageStartY+imageHeight/2)
		pdf.CellFormat(contentWidth, 10, "No Image Available", "", 0, "C", false, 0, "")
	}
	
	// Property Title (Use Arabic localized title if available)
	pdf.SetY(186)
	if s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 24)
	} else {
		pdf.SetFont("Arial", "B", 26)
	}
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	
	// Use localized Arabic title if available, otherwise fallback to English title
	title := property.Title
	if property.ArabicContent.Title != "" {
		title = property.ArabicContent.Title
		title = s.fixMojibakeLatin1ToUTF8(title)
	}
	
	titleLines := pdf.SplitLines([]byte(title), contentWidth)
	for _, line := range titleLines {
		pdf.CellFormat(contentWidth, 12, string(line), "", 1, "C", false, 0, "")
	}
	pdf.Ln(3)
	
	// Add a subtle price background box for emphasis
	priceBoxY := pdf.GetY()
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(marginX+35, priceBoxY-2, contentWidth-70, 18, "F")
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Rect(marginX+35, priceBoxY-2, contentWidth-70, 18, "D")
	
	// Price (prominent, gold color)
	pdf.SetY(priceBoxY)
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(goldR, goldG, goldB)
	priceText := s.formatPrice(property.Price, property.Currency)
	pdf.CellFormat(contentWidth, 14, priceText, "", 1, "C", false, 0, "")
	pdf.Ln(5)
	
	// Location (gray, medium size)
	pdf.SetFont("Arial", "", 13)
	pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
	locationText := s.formatLocation(property)
	pdf.MultiCell(contentWidth, 6, locationText, "", "C", false)
	
	// Decorative bottom section with elegant design
	pdf.SetY(268)
	
	// Add decorative diamond shape in center
	centerX := pageWidth / 2
	diamondY := 272.0
	pdf.SetFillColor(goldR, goldG, goldB)
	// Create diamond with lines
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Line(centerX-4, diamondY, centerX, diamondY-3)
	pdf.Line(centerX, diamondY-3, centerX+4, diamondY)
	pdf.Line(centerX+4, diamondY, centerX, diamondY+3)
	pdf.Line(centerX, diamondY+3, centerX-4, diamondY)
	
	// Lines extending from diamond
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX+50, diamondY, centerX-6, diamondY)
	pdf.Line(centerX+6, diamondY, pageWidth-marginX-50, diamondY)
	
	s.addPageNumber(pdf, 1)
}

// addDetailsPageArabicCombined creates the Arabic property description, highlights, amenities, investment opportunity, and gallery
func (s *PDFService) addDetailsPageArabicCombined(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
	
	// Add cream background
	s.addPageBackground(pdf)
	
	s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	// Use localized content if available, fallback to legacy
	var descLabel, highlightsLabel, amenitiesLabel string
	var description string
	var highlights []string
	var amenities []string
	
	if property.ArabicContent.Description != "" {
		// Use new localized content
		descLabel = property.ArabicContent.PropertyDescriptionLabel
		highlightsLabel = property.ArabicContent.KeyHighlightsLabel
		amenitiesLabel = property.ArabicContent.AmenitiesLabel
		description = property.ArabicContent.Description
		highlights = property.ArabicContent.Highlights
		amenities = property.ArabicContent.Amenities
	} else {
		// Fallback to legacy fields
		descLabel = "وصف العقار"
		highlightsLabel = "المميزات الرئيسية"
		amenitiesLabel = "المرافق والميزات"
		description = property.AIContent.ArabicDescription
		highlights = []string{} // Legacy didn't have Arabic highlights
		amenities = property.Amenities
	}
	
	if description == "" {
		description = "لا يوجد وصف متاح"
	}
	
	// Section: Arabic Description
	if s.hasArabicFont {
		currentY = s.addSectionHeaderAligned(pdf, descLabel, currentY, s.arabicFontName, "R")
	} else {
		currentY = s.addSectionHeader(pdf, descLabel, currentY)
	}
	
	// Use Arabic font if available
	if s.hasArabicFont {
		pdf.SetFont(s.arabicFontName, "", 12)
	} else {
		pdf.SetFont("Arial", "", 11)
	}
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, currentY)
	
	// Right-aligned for Arabic text
	description = s.fixMojibakeLatin1ToUTF8(description)
	pdf.MultiCell(contentWidth, 6, description, "", "R", false)
	currentY = pdf.GetY() + 8
	
	// Section: Key Highlights (Arabic)
	if len(highlights) > 0 {
		if currentY > 220 {
			pdf.AddPage()
			s.addPageBackground(pdf)
			s.addBrandingIfAvailable(pdf)
			currentY = marginY + 10
		}
		
		if s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, highlightsLabel, currentY, s.arabicFontName, "R")
		} else {
			currentY = s.addSectionHeader(pdf, highlightsLabel, currentY)
		}
		
		if s.hasArabicFont {
			pdf.SetFont(s.arabicFontName, "", 11)
		} else {
			pdf.SetFont("Arial", "", 11)
		}
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
		for _, raw := range highlights {
			highlight := s.sanitizeBulletText(raw)
			highlight = s.fixMojibakeLatin1ToUTF8(highlight)
			
			// Draw a gold bullet (filled circle)
			bulletX := pageWidth - marginX - 5 // Right side for RTL
			bulletY := currentY + 3.5
			pdf.SetFillColor(goldR, goldG, goldB)
			pdf.Circle(bulletX, bulletY, 1.6, "F")
			
			// Highlight text (right-aligned)
			pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
			if s.hasArabicFont {
				pdf.SetFont(s.arabicFontName, "", 11)
			} else {
				pdf.SetFont("Arial", "", 11)
			}
			pdf.SetXY(marginX, currentY)
			pdf.MultiCell(contentWidth-12, 6, highlight, "", "R", false)
			currentY = pdf.GetY() + 1
		}
		currentY += 6
	}
	
	// Section: Amenities (if available)
	if len(amenities) > 0 {
		if currentY > 220 {
			pdf.AddPage()
			s.addPageBackground(pdf)
			s.addBrandingIfAvailable(pdf)
			currentY = marginY + 10
		}
		
		if s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, amenitiesLabel, currentY, s.arabicFontName, "R")
		} else {
			currentY = s.addSectionHeader(pdf, amenitiesLabel, currentY)
		}
		
		if s.hasArabicFont {
			pdf.SetFont(s.arabicFontName, "", 10)
		} else {
			pdf.SetFont("Arial", "", 10)
		}
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
		// Display amenities in a 2-column grid with checkmarks
		colWidth := (contentWidth - 10) / 2
		amenityHeight := 7.0
		
		for i, amenity := range amenities {
			col := i % 2
			xPos := marginX + float64(col)*(colWidth+10)
			
			pdf.SetXY(xPos, currentY)
			
			// Draw a green check mark using vector lines
			pdf.SetDrawColor(46, 125, 50)
			pdf.SetLineWidth(0.8)
			startX := xPos
			startY := currentY + amenityHeight/2
			pdf.Line(startX, startY, startX+2.0, startY+2.0)
			pdf.Line(startX+2.0, startY+2.0, startX+6.0, startY-1.0)
			
			// Amenity text (apply mojibake fix for Arabic)
			amenity = s.fixMojibakeLatin1ToUTF8(amenity)
			pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
			if s.hasArabicFont {
				pdf.SetFont(s.arabicFontName, "", 10)
			} else {
				pdf.SetFont("Arial", "", 10)
			}
			pdf.SetX(xPos + 9)
			pdf.CellFormat(colWidth-7, amenityHeight, amenity, "", 0, "", false, 0, "")
			
			// Move to next row after 2 columns
			if col == 1 {
				currentY += amenityHeight
			}
		}
		
		// Handle odd number of amenities
		if len(amenities)%2 == 1 {
			currentY += amenityHeight
		}
	}
	
	currentY += 8
	
	// Section: Additional Content (Investment Opportunity) - Arabic
	var additionalTitle, additionalContent string
	if property.ArabicContent.AdditionalSectionTitle != "" {
		additionalTitle = property.ArabicContent.AdditionalSectionTitle
		additionalContent = property.ArabicContent.AdditionalSectionContent
	} else {
		additionalTitle = "فرصة استثمارية"
		additionalContent = "يمثل هذا العقار فرصة استثمارية ممتازة في موقع متميز."
	}
	
	// Check if we need a new page for investment content
	if currentY > 200 {
		pdf.AddPage()
		s.addPageBackground(pdf)
		s.addBrandingIfAvailable(pdf)
		currentY = marginY + 10
	}
	
	if additionalContent != "" {
		if s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, additionalTitle, currentY, s.arabicFontName, "R")
		} else {
			currentY = s.addSectionHeader(pdf, additionalTitle, currentY)
		}
		
		if s.hasArabicFont {
			pdf.SetFont(s.arabicFontName, "", 11)
		} else {
			pdf.SetFont("Arial", "", 10.5)
		}
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		pdf.SetXY(marginX, currentY)
		additionalContent = s.fixMojibakeLatin1ToUTF8(additionalContent)
		pdf.MultiCell(contentWidth, 5.5, additionalContent, "", "R", false)
		currentY = pdf.GetY() + 8
	}
	
	// Add Property Gallery (if images available) on the same page
	if len(property.ImageURLs) > 1 {
		// Check if we need a new page for gallery
		if currentY > 200 {
			pdf.AddPage()
			s.addPageBackground(pdf)
			s.addBrandingIfAvailable(pdf)
			currentY = marginY + 10
		}
		
		galleryLabel := "معرض العقار"
		if property.ArabicContent.PropertyGalleryLabel != "" {
			galleryLabel = property.ArabicContent.PropertyGalleryLabel
		}
		galleryLabel = s.fixMojibakeLatin1ToUTF8(galleryLabel)
		
		if s.hasArabicFont {
			currentY = s.addSectionHeaderAligned(pdf, galleryLabel, currentY, s.arabicFontName, "R")
		} else {
			currentY = s.addSectionHeader(pdf, galleryLabel, currentY)
		}
		currentY += 3
		
		// Display up to 4 additional images in a compact 2x2 grid
		imgWidth := (contentWidth - 8) / 2
		imgHeight := imgWidth * 0.65
		spacing := 8.0
		
		imageCount := 0
		maxImages := 4
		
		for i := 1; i < len(property.ImageURLs) && imageCount < maxImages; i++ {
			row := imageCount / 2
			col := imageCount % 2
			
			xPos := marginX + float64(col)*(imgWidth+spacing)
			yPos := currentY + float64(row)*(imgHeight+spacing)
			
			// Check if we're running out of space
			if yPos+imgHeight > pageHeight-25 {
				break
			}
			
			// Add shadow effect
			pdf.SetFillColor(180, 180, 180)
			pdf.Rect(xPos+1.5, yPos+1.5, imgWidth, imgHeight, "F")
			
			// Add white background
			pdf.SetFillColor(255, 255, 255)
			pdf.Rect(xPos, yPos, imgWidth, imgHeight, "F")
			
			// Add gold border/frame effect
			pdf.SetDrawColor(goldR, goldG, goldB)
			pdf.SetLineWidth(0.6)
			pdf.Rect(xPos, yPos, imgWidth, imgHeight, "D")
			
			err := s.addImageFromURL(pdf, property.ImageURLs[i], xPos+2, yPos+2, imgWidth-4, imgHeight-4)
			if err != nil {
				// Placeholder for failed images
				pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
				pdf.Rect(xPos+2, yPos+2, imgWidth-4, imgHeight-4, "F")
			}
			
			imageCount++
		}
	}
	
	// Add decorative bottom diamond element
	s.addBottomDiamondDecoration(pdf)
	
	s.addPageNumber(pdf, 2)
}

