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

// Color scheme constants
const (
	// Primary colors
	darkBlueR, darkBlueG, darkBlueB = 31, 78, 121   // #1F4E79 - Headers
	goldR, goldG, goldB             = 212, 175, 55  // #D4AF37 - Accents
	
	// Secondary colors
	lightGrayR, lightGrayG, lightGrayB = 245, 245, 245 // #F5F5F5 - Backgrounds
	darkGrayR, darkGrayG, darkGrayB    = 60, 60, 60    // #3C3C3C - Body text
	mediumGrayR, mediumGrayG, mediumGrayB = 120, 120, 120 // #787878 - Secondary text
	
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
	pdf.SetAutoPageBreak(false, 15) // Disable auto page break for better control
    s.setupFonts(pdf)
	
	// Page 1: Cover Page
	s.addCoverPage(pdf, property)
	
	// Page 2: Property Description & Details
	s.addDetailsPage(pdf, property)
	
	// Page 3: Additional Images (if available)
	if len(property.ImageURLs) > 1 {
		s.addGalleryPage(pdf, property)
	}
	
	// Page 4: Arabic Description & Agent Info
	s.addArabicAndContactPage(pdf, property)
	
	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addCoverPage creates an attractive cover page with main image, title, and price
func (s *PDFService) addCoverPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
    s.addBrandingIfAvailable(pdf)
	
	// Add gold accent bar at top
	pdf.SetFillColor(goldR, goldG, goldB)
	pdf.Rect(0, 0, pageWidth, 8, "F")
	
	// Add main property image (large, full-width)
	imageHeight := 160.0
	if len(property.ImageURLs) > 0 {
		// Add image with slight margins
		err := s.addImageFromURL(pdf, property.ImageURLs[0], marginX, 20, contentWidth, imageHeight)
		if err != nil {
			// If image fails, create a placeholder
			pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
			pdf.Rect(marginX, 20, contentWidth, imageHeight, "F")
			pdf.SetFont("Arial", "I", 12)
			pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
			pdf.SetXY(marginX, 20+imageHeight/2)
			pdf.CellFormat(contentWidth, 10, "Image Not Available", "", 0, "C", false, 0, "")
		}
	} else {
		// Placeholder for missing image
		pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
		pdf.Rect(marginX, 20, contentWidth, imageHeight, "F")
		pdf.SetFont("Arial", "I", 12)
		pdf.SetTextColor(mediumGrayR, mediumGrayG, mediumGrayB)
		pdf.SetXY(marginX, 20+imageHeight/2)
		pdf.CellFormat(contentWidth, 10, "No Image Available", "", 0, "C", false, 0, "")
	}
	
	// Property Title (large, bold, dark blue)
	pdf.SetY(190)
	pdf.SetFont("Arial", "B", 26)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	
	// Handle long titles
	titleLines := pdf.SplitLines([]byte(property.Title), contentWidth)
	for _, line := range titleLines {
		pdf.CellFormat(contentWidth, 12, string(line), "", 1, "C", false, 0, "")
	}
	pdf.Ln(3)
	
	// Price (prominent, gold color)
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
	
	// Decorative bottom line
	pdf.SetY(270)
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginX+50, 270, pageWidth-marginX-50, 270)
	
	// Add page number
	s.addPageNumber(pdf, 1)
}

// addDetailsPage creates the property description, highlights, and amenities page
func (s *PDFService) addDetailsPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	// Section: Property Description
	currentY = s.addSectionHeader(pdf, "Property Description", currentY)
	
    if s.hasBodyFont {
        pdf.SetFont(s.bodyFontName, "", 11)
    } else {
        pdf.SetFont("Arial", "", 11)
    }
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX, currentY)
	
	description := property.AIContent.EnglishDescription
	if description == "" {
		description = property.Description
	}
	if description == "" {
		description = "No description available."
	}
	
	pdf.MultiCell(contentWidth, 5.5, description, "", "L", false)
	currentY = pdf.GetY() + 8
	
    // Section: Key Highlights
	if len(property.AIContent.KeyHighlights) > 0 {
		currentY = s.addSectionHeader(pdf, "Key Highlights", currentY)
		
		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
        for _, raw := range property.AIContent.KeyHighlights {
            highlight := s.sanitizeBulletText(raw)
            // Draw a gold bullet (filled circle) to avoid Unicode bullet issues
            bulletX := marginX + 5
            bulletY := currentY + 3.5
            pdf.SetFillColor(goldR, goldG, goldB)
            pdf.Circle(bulletX, bulletY, 1.6, "F")

            // Highlight text
            pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
            pdf.SetFont("Arial", "", 11)
            pdf.SetXY(marginX+12, currentY)
            pdf.MultiCell(contentWidth-12, 6, highlight, "", "L", false)
            currentY = pdf.GetY() + 1
        }
		currentY += 6
	}
	
	// Section: Amenities
	if len(property.Amenities) > 0 {
		// Check if we need a new page
		if currentY > 220 {
			pdf.AddPage()
			currentY = marginY + 10
		}
		
		currentY = s.addSectionHeader(pdf, "Amenities & Features", currentY)
		
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
		
        // Display amenities in a 2-column grid with checkmarks
		colWidth := (contentWidth - 10) / 2
		amenityHeight := 7.0
		
        for i, amenity := range property.Amenities {
			col := i % 2
			xPos := marginX + float64(col)*(colWidth+10)
			
			pdf.SetXY(xPos, currentY)
			
            // Draw a green check mark using vector lines (avoids Unicode glyph issues)
            pdf.SetDrawColor(46, 125, 50)
            pdf.SetLineWidth(0.8)
            startX := xPos
            startY := currentY + amenityHeight/2
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
				currentY += amenityHeight
			}
		}
		
		// Handle odd number of amenities
		if len(property.Amenities)%2 == 1 {
			currentY += amenityHeight
		}
	}
	
	// Add page number
	s.addPageNumber(pdf, 2)
}

// addGalleryPage creates an image gallery for additional property photos
func (s *PDFService) addGalleryPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
    s.addBrandingIfAvailable(pdf)
	currentY := marginY + 10.0
	
	// Section header
	currentY = s.addSectionHeader(pdf, "Property Gallery", currentY)
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
		
		// Add border/frame effect
		pdf.SetDrawColor(mediumGrayR, mediumGrayG, mediumGrayB)
		pdf.SetLineWidth(0.2)
		pdf.Rect(xPos, yPos, imgWidth, imgHeight, "D")
		
		err := s.addImageFromURL(pdf, property.ImageURLs[i], xPos+1, yPos+1, imgWidth-2, imgHeight-2)
		if err != nil {
			// Placeholder for failed images
			pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
			pdf.Rect(xPos+1, yPos+1, imgWidth-2, imgHeight-2, "F")
		}
		
		imageCount++
	}
	
	// Add page number
	s.addPageNumber(pdf, 3)
}

// addArabicAndContactPage creates the Arabic description and agent contact page
func (s *PDFService) addArabicAndContactPage(pdf *gofpdf.Fpdf, property *models.Property) {
	pdf.AddPage()
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
	
	// Agent Contact Card
	s.addAgentContactCard(pdf, property, currentY)
	
	// Add page number
	pageNum := 4
	if len(property.ImageURLs) <= 1 {
		pageNum = 3
	}
	s.addPageNumber(pdf, pageNum)
}

// addAgentContactCard creates a professional contact card for the agent
func (s *PDFService) addAgentContactCard(pdf *gofpdf.Fpdf, property *models.Property, startY float64) {
	cardHeight := 55.0
	cardY := pageHeight - marginY - cardHeight - 20
	
	// Background card
	pdf.SetFillColor(lightGrayR, lightGrayG, lightGrayB)
	pdf.Rect(marginX, cardY, contentWidth, cardHeight, "F")
	
	// Gold accent border
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.8)
	pdf.Rect(marginX, cardY, contentWidth, cardHeight, "D")
	
	// "Contact Agent" header
	pdf.SetXY(marginX+5, cardY+5)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.CellFormat(contentWidth-10, 8, "Contact Your Agent", "", 1, "C", false, 0, "")
	
	// Divider line
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.3)
	pdf.Line(marginX+30, cardY+13, pageWidth-marginX-30, cardY+13)
	
	// Agent info
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, cardY+18)
	pdf.CellFormat(25, 6, "Name:", "", 0, "", false, 0, "")
            if s.hasBodyFont {
                pdf.SetFont(s.bodyFontName, "", 11)
            } else {
                pdf.SetFont("Arial", "", 11)
            }
	pdf.CellFormat(0, 6, property.AgentInfo.Name, "", 0, "", false, 0, "")
	
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(marginX+10, cardY+28)
	pdf.CellFormat(25, 6, "Email:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.CellFormat(0, 6, property.AgentInfo.Email, "", 0, "", false, 0, "")
	
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetXY(marginX+10, cardY+38)
	pdf.CellFormat(25, 6, "Phone:", "", 0, "", false, 0, "")
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
    // Optional: allow specifying a TTF for Arabic via env var
    if fontPath := os.Getenv("ARABIC_TTF_PATH"); fontPath != "" {
        if _, err := os.Stat(fontPath); err == nil {
            pdf.AddUTF8Font("ArabicFont", "", fontPath)
            s.arabicFontName = "ArabicFont"
            s.hasArabicFont = true
            fmt.Println("[PDF] Loaded Arabic UTF-8 font:", fontPath)
        } else {
            fmt.Println("[PDF] ARABIC_TTF_PATH not found:", fontPath, "err:", err)
        }
    } else {
        fmt.Println("[PDF] ARABIC_TTF_PATH not set; Arabic text may render incorrectly.")
    }

    // Optional: general body font for all content (Unicode)
    if bodyPath := os.Getenv("BODY_TTF_PATH"); bodyPath != "" {
        if _, err := os.Stat(bodyPath); err == nil {
            pdf.AddUTF8Font("BodyFont", "", bodyPath)
            s.bodyFontName = "BodyFont"
            s.hasBodyFont = true
            fmt.Println("[PDF] Loaded Body UTF-8 font:", bodyPath)
        } else {
            fmt.Println("[PDF] BODY_TTF_PATH not found:", bodyPath, "err:", err)
        }
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

