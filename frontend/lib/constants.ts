
// API Configuration
export const API_CONFIG = {
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000',
  endpoints: {
    submitProperty: '/api/property',
  },
  timeout: 30000, // 30 seconds
}

// Currency Configuration
export const CURRENCY_SYMBOLS = {
  Rupees: '₹',
  Dollar: '$',
  Dirhams: 'د.إ',
} as const;

// Form Configuration
export const FORM_CONFIG = {
  maxImages: 10,
  maxImageSize: 10 * 1024 * 1024, // 10MB in bytes
  acceptedImageTypes: ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'],
  minDescriptionLength: 50,
};

// UI Configuration
export const UI_CONFIG = {
  successMessageDuration: 3000, // milliseconds
  submitSimulationDelay: 2000, // milliseconds
};

// Validation Patterns
export const VALIDATION_PATTERNS = {
  zipCode: /^\d{5,6}(-\d{4})?$/,
  email: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
  phone: /^[\d\s\-\(\)]+$/,
};

// Suggested Amenities (for future autocomplete)
export const COMMON_AMENITIES = [
  'Swimming Pool',
  'Gym/Fitness Center',
  'Parking',
  'Balcony',
  'Garden',
  'Security System',
  'Central Air Conditioning',
  'Hardwood Floors',
  'Updated Kitchen',
  'Walk-in Closet',
  'Fireplace',
  'Pet Friendly',
  'Washer/Dryer',
  'Dishwasher',
  'Storage Space',
];

