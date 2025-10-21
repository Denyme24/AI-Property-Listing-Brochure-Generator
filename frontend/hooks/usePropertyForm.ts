import { useState } from "react";
import type { ImagePreview } from "@/types/property";

export function usePropertyForm() {
  const [images, setImages] = useState<ImagePreview[]>([]);
  const [amenities, setAmenities] = useState<string[]>([]);
  const [amenityInput, setAmenityInput] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleImageUpload = (files: FileList | null) => {
    if (!files) return;
    
    const fileArray = Array.from(files);
    const newImages: ImagePreview[] = fileArray.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));
    
    setImages((prev) => [...prev, ...newImages]);
  };

  const removeImage = (index: number) => {
    setImages((prev) => {
      const newImages = [...prev];
      // Clean up object URL to prevent memory leaks
      URL.revokeObjectURL(newImages[index].preview);
      newImages.splice(index, 1);
      return newImages;
    });
  };

  const addAmenity = (value?: string) => {
    const amenityValue = value || amenityInput;
    if (amenityValue.trim() && !amenities.includes(amenityValue.trim())) {
      setAmenities((prev) => [...prev, amenityValue.trim()]);
      setAmenityInput("");
    }
  };

  const removeAmenity = (amenity: string) => {
    setAmenities((prev) => prev.filter((a) => a !== amenity));
  };


  const resetFormState = () => {
    // Clean up all object URLs
    images.forEach((img) => URL.revokeObjectURL(img.preview));
    
    setImages([]);
    setAmenities([]);
    setAmenityInput("");
    setIsSubmitting(false);
    setIsSuccess(false);
  };

  return {
    // State
    images,
    amenities,
    amenityInput,
    isSubmitting,
    isSuccess,
    
    // Setters
    setAmenityInput,
    setIsSubmitting,
    setIsSuccess,
    
    // Actions
    handleImageUpload,
    removeImage,
    addAmenity,
    removeAmenity,
    resetFormState,
  };
}

