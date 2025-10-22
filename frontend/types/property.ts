export type Currency = "Rupees" | "Dollar" | "Dirhams";

export interface PropertyFormData {
  // Property Information
  title: string;
  description: string;
  price: number;
  currency: Currency;
  
  // Location Details
  address: string;
  city: string;
  state: string;
  zipCode: string;
  
  // Features
  amenities: string[];
  
  // Images
  images: File[];
  
  // Agent Information
  agentName: string;
  agentEmail: string;
  agentPhone: string;
}

export interface PropertySubmissionResponse {
  success: boolean;
  message: string;
  propertyId?: string;
  pdfUrl?: string; // Legacy field
  pdfViewUrl?: string; // Legacy field
  pdfDownloadUrl?: string; // Legacy field
  pdfUrlEnglish?: string;
  pdfUrlArabic?: string;
  pdfViewUrlEnglish?: string;
  pdfViewUrlArabic?: string;
  pdfDownloadUrlEnglish?: string;
  pdfDownloadUrlArabic?: string;
}

export interface ImagePreview {
  file: File;
  preview: string;
}

