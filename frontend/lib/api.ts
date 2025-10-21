import type { PropertyFormData, PropertySubmissionResponse } from "@/types/property";
import { API_CONFIG } from "./constants";

export async function submitPropertyListing(
  data: PropertyFormData,
  amenities: string[],
  images: File[]
): Promise<PropertySubmissionResponse> {
  try {
    // Validate required fields
    if (!data.title || !data.price || !data.address || !data.city || !data.state || !data.zipCode) {
      throw new Error('Please fill in all required fields');
    }

    if (images.length === 0) {
      throw new Error('Please upload at least one image');
    }

    // Create FormData for file upload
    const formData = new FormData();

    // Append property data
    formData.append('title', data.title);
    formData.append('description', data.description || '');
    formData.append('price', data.price.toString());
    formData.append('currency', data.currency || 'Dollar');
    formData.append('address', data.address);
    formData.append('city', data.city);
    formData.append('state', data.state);
    formData.append('zipCode', data.zipCode);
    formData.append('agentName', data.agentName);
    formData.append('agentEmail', data.agentEmail);
    formData.append('agentPhone', data.agentPhone);

    // Append amenities
    amenities.forEach((amenity) => {
      formData.append('amenities[]', amenity);
    });

    // Append images
    images.forEach((image) => {
      formData.append('images[]', image);
    });

    // Make API request
    const response = await fetch(
      `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.submitProperty}`,
      {
        method: 'POST',
        body: formData,
      }
    );

    const result = await response.json();

    if (!response.ok) {
      throw new Error(result.message || result.error || `HTTP error! status: ${response.status}`);
    }

    return result;

  } catch (error) {
    console.error('Error submitting property:', error);
    throw error;
  }
}

export async function submitPropertyListingJSON(
  data: PropertyFormData,
  amenities: string[],
  imageUrls: string[]
): Promise<PropertySubmissionResponse> {
  try {
    const payload = {
      ...data,
      amenities,
      imageUrls,
    };

    const response = await fetch(
      `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.submitProperty}`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result: PropertySubmissionResponse = await response.json();
    return result;

  } catch (error) {
    console.error('Error submitting property:', error);
    throw error;
  }
}

export async function downloadPDF(pdfUrl: string, filename: string = 'property-brochure.pdf') {
  try {
    // Pre-signed URLs can be opened directly without fetching
    // This avoids CORS issues with S3
    const link = document.createElement('a');
    link.href = pdfUrl;
    link.download = filename;
    link.target = '_blank'; // Open in new tab as fallback
    link.rel = 'noopener noreferrer';
    
    document.body.appendChild(link);
    link.click();
    
    // Clean up
    setTimeout(() => {
      document.body.removeChild(link);
    }, 100);
  } catch (error) {
    console.error('Error downloading PDF:', error);
    throw error;
  }
}

