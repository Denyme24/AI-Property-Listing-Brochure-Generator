package services

import (
	"bytes"
	"fmt"
	"net/http"
	"property-brochure-backend/models"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateBrochure(property *models.Property) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 10)

	// Add fonts for Arabic support (using built-in fonts for basic support)
	pdf.AddPage()

	// Header - Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(31, 78, 121) // Dark blue
	pdf.MultiCell(0, 12, property.Title, "", "C", false)
	pdf.Ln(5)

	// Price
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(220, 53, 69) // Red
	priceText := fmt.Sprintf("%s %.2f", property.Currency, property.Price)
	pdf.CellFormat(0, 10, priceText, "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Location
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(100, 100, 100)
	locationText := fmt.Sprintf("%s, %s, %s %s", property.Address, property.City, property.State, property.ZipCode)
	pdf.MultiCell(0, 6, locationText, "", "C", false)
	pdf.Ln(5)

	// Line separator
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(8)

	// Add main image if available
	if len(property.ImageURLs) > 0 {
		if err := s.addImageFromURL(pdf, property.ImageURLs[0], 20, pdf.GetY(), 170, 100); err == nil {
			pdf.Ln(105)
		}
	}

	// English Description Section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(31, 78, 121)
	pdf.Cell(0, 10, "Property Description")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(60, 60, 60)
	pdf.MultiCell(0, 6, property.AIContent.EnglishDescription, "", "L", false)
	pdf.Ln(5)

	// Key Highlights Section
	if len(property.AIContent.KeyHighlights) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(31, 78, 121)
		pdf.Cell(0, 10, "Key Highlights")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(60, 60, 60)
		for _, highlight := range property.AIContent.KeyHighlights {
			pdf.Cell(10, 6, "")
			pdf.Cell(0, 6, "• "+highlight)
			pdf.Ln(6)
		}
		pdf.Ln(3)
	}

	// Amenities Section
	if len(property.Amenities) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(31, 78, 121)
		pdf.Cell(0, 10, "Amenities")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(60, 60, 60)
		
		// Display amenities in two columns
		amenitiesPerCol := (len(property.Amenities) + 1) / 2
		for i := 0; i < amenitiesPerCol; i++ {
			// Left column
			if i < len(property.Amenities) {
				pdf.Cell(10, 6, "")
				pdf.Cell(85, 6, "✓ "+property.Amenities[i])
			}
			
			// Right column
			rightIndex := i + amenitiesPerCol
			if rightIndex < len(property.Amenities) {
				pdf.Cell(10, 6, "")
				pdf.Cell(85, 6, "✓ "+property.Amenities[rightIndex])
			}
			pdf.Ln(6)
		}
		pdf.Ln(3)
	}

	// Additional images (if any)
	if len(property.ImageURLs) > 1 {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(31, 78, 121)
		pdf.Cell(0, 10, "Additional Photos")
		pdf.Ln(10)

		x, y := 20.0, pdf.GetY()
		imgWidth := 80.0
		imgHeight := 60.0
		spacing := 10.0

		for i := 1; i < len(property.ImageURLs) && i < 5; i++ {
			if err := s.addImageFromURL(pdf, property.ImageURLs[i], x, y, imgWidth, imgHeight); err == nil {
				if (i % 2) == 1 {
					x += imgWidth + spacing
				} else {
					x = 20
					y += imgHeight + spacing
					if y > 250 {
						break
					}
				}
			}
		}
	}

	// Arabic Description Section (on new page)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(31, 78, 121)
	pdf.Cell(0, 10, "Arabic Description | الوصف بالعربية")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(60, 60, 60)
	// Note: For proper Arabic rendering, you'd need to use a library that supports RTL text
	// For now, we'll display it as-is
	pdf.MultiCell(0, 6, property.AIContent.ArabicDescription, "", "R", false)
	pdf.Ln(10)

	// Agent Information Section (Footer)
	pdf.SetY(-40)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(31, 78, 121)
	pdf.Cell(0, 6, "Contact Agent")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(60, 60, 60)
	pdf.Cell(0, 5, "Name: "+property.AgentInfo.Name)
	pdf.Ln(5)
	pdf.Cell(0, 5, "Email: "+property.AgentInfo.Email)
	pdf.Ln(5)
	pdf.Cell(0, 5, "Phone: "+property.AgentInfo.Phone)

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
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

	// Determine image type from content type
	imageType := "jpg"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "png") {
		imageType = "png"
	} else if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		imageType = "jpg"
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
	
	pdf.RegisterImageOptionsReader(uniqueName, opts, resp.Body)
	pdf.ImageOptions(uniqueName, x, y, w, h, false, opts, 0, "")

	return nil
}

