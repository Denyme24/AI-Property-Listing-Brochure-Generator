"use client";

import React from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { X, Upload, Loader2, CheckCircle2, Download } from "lucide-react";
import Image from "next/image";
import type { PropertyFormData, Currency } from "@/types/property";
import { usePropertyForm } from "@/hooks/usePropertyForm";
import { submitPropertyListing } from "@/lib/api";
import { toast } from "sonner";

interface PropertyFormProps {
  onClose?: () => void;
}

export function PropertyForm({ onClose }: PropertyFormProps) {
  const {
    images,
    amenities,
    amenityInput,
    isSubmitting,
    isSuccess,
    setAmenityInput,
    setIsSubmitting,
    setIsSuccess,
    handleImageUpload,
    removeImage,
    addAmenity,
    removeAmenity,
    resetFormState,
  } = usePropertyForm();

  const [pdfViewUrlEnglish, setPdfViewUrlEnglish] = React.useState<
    string | null
  >(null);
  const [pdfDownloadUrlEnglish, setPdfDownloadUrlEnglish] = React.useState<
    string | null
  >(null);
  const [pdfViewUrlArabic, setPdfViewUrlArabic] = React.useState<string | null>(
    null
  );
  const [pdfDownloadUrlArabic, setPdfDownloadUrlArabic] = React.useState<
    string | null
  >(null);
  const [propertyId, setPropertyId] = React.useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<PropertyFormData>({
    defaultValues: {
      currency: "Dollar",
    },
  });

  // Handle form submission
  const onSubmit = async (data: PropertyFormData) => {
    // Validate images
    if (images.length === 0) {
      toast.error("Please upload at least one image");
      return;
    }

    // Validate amenities
    if (amenities.length === 0) {
      toast.error("Please add at least one amenity");
      return;
    }

    setIsSubmitting(true);

    try {
      // Submit property listing
      const imageFiles = images.map((img) => img.file);
      const response = await submitPropertyListing(data, amenities, imageFiles);

      if (response.success) {
        setIsSuccess(true);
        setPdfViewUrlEnglish(
          response.pdfViewUrlEnglish || response.pdfUrlEnglish || null
        );
        setPdfDownloadUrlEnglish(
          response.pdfDownloadUrlEnglish || response.pdfUrlEnglish || null
        );
        setPdfViewUrlArabic(
          response.pdfViewUrlArabic || response.pdfUrlArabic || null
        );
        setPdfDownloadUrlArabic(
          response.pdfDownloadUrlArabic || response.pdfUrlArabic || null
        );
        setPropertyId(response.propertyId || null);
        toast.success(
          response.message || "Property listing created successfully!"
        );
      } else {
        throw new Error(response.message || "Failed to submit property");
      }
    } catch (error) {
      const errorMsg =
        error instanceof Error ? error.message : "An unexpected error occurred";
      toast.error(errorMsg);
      console.error("Error submitting property:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDownloadPDF = (language: "english" | "arabic") => {
    const url =
      language === "english" ? pdfDownloadUrlEnglish : pdfDownloadUrlArabic;
    if (!url) return;

    // Open the download URL directly - it will force download due to attachment disposition
    window.open(url, "_blank");
    toast.success(
      `PDF download started (${language === "english" ? "English" : "Arabic"})!`
    );
  };

  const handleViewPDF = (language: "english" | "arabic") => {
    const url = language === "english" ? pdfViewUrlEnglish : pdfViewUrlArabic;
    if (!url) return;

    // Open the view URL in a new tab - it will display inline in browser
    window.open(url, "_blank");
  };

  const handleReset = () => {
    reset();
    resetFormState();
    setPdfViewUrlEnglish(null);
    setPdfDownloadUrlEnglish(null);
    setPdfViewUrlArabic(null);
    setPdfDownloadUrlArabic(null);
    setPropertyId(null);
  };

  if (isSuccess) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-6 text-center">
        <div className="h-20 w-20 bg-green-100 rounded-full flex items-center justify-center">
          <CheckCircle2 className="h-12 w-12 text-green-600" />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-bold text-gray-900">
            Brochure Generated Successfully!
          </h2>
          <p className="text-gray-600 max-w-md">
            Your property brochure has been generated with AI-powered content in
            both English and Arabic.
          </p>
          {propertyId && (
            <p className="text-sm text-gray-500">
              Property ID: <span className="font-mono">{propertyId}</span>
            </p>
          )}
        </div>
        <div className="flex flex-col gap-4 w-full max-w-2xl">
          {/* English Brochure */}
          {(pdfViewUrlEnglish || pdfDownloadUrlEnglish) && (
            <div className="border border-gray-200 rounded-lg p-4 bg-white">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-gray-900">
                  English Brochure
                </h3>
                <Badge
                  variant="outline"
                  className="bg-blue-50 text-blue-700 border-blue-200"
                >
                  English
                </Badge>
              </div>
              <div className="flex gap-3">
                {pdfViewUrlEnglish && (
                  <Button
                    onClick={() => handleViewPDF("english")}
                    variant="outline"
                    className="flex-1 gap-2"
                  >
                    View PDF
                  </Button>
                )}
                {pdfDownloadUrlEnglish && (
                  <Button
                    onClick={() => handleDownloadPDF("english")}
                    className="flex-1 gap-2"
                  >
                    <Download className="h-4 w-4" />
                    Download PDF
                  </Button>
                )}
              </div>
            </div>
          )}

          {/* Arabic Brochure */}
          {(pdfViewUrlArabic || pdfDownloadUrlArabic) && (
            <div className="border border-gray-200 rounded-lg p-4 bg-white">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-gray-900">Arabic Brochure</h3>
                <Badge
                  variant="outline"
                  className="bg-green-50 text-green-700 border-green-200"
                >
                  العربية
                </Badge>
              </div>
              <div className="flex gap-3">
                {pdfViewUrlArabic && (
                  <Button
                    onClick={() => handleViewPDF("arabic")}
                    variant="outline"
                    className="flex-1 gap-2"
                  >
                    View PDF
                  </Button>
                )}
                {pdfDownloadUrlArabic && (
                  <Button
                    onClick={() => handleDownloadPDF("arabic")}
                    className="flex-1 gap-2"
                  >
                    <Download className="h-4 w-4" />
                    Download PDF
                  </Button>
                )}
              </div>
            </div>
          )}

          <div className="flex gap-3 justify-center pt-2">
            <Button onClick={handleReset} variant="outline">
              Create Another Listing
            </Button>
            {onClose && (
              <Button onClick={onClose} variant="ghost">
                Close
              </Button>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
      {/* Property Information */}
      <Card>
        <CardHeader>
          <CardTitle>Property Information</CardTitle>
          <CardDescription>
            Enter the basic details of your property listing
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="title">Property Title *</Label>
            <Input
              id="title"
              placeholder="e.g., Stunning 3BR Modern Home in Downtown"
              {...register("title", { required: "Property title is required" })}
              className={errors.title ? "border-red-500" : ""}
            />
            {errors.title && (
              <p className="text-sm text-red-500">{errors.title.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description *</Label>
            <Textarea
              id="description"
              placeholder="Describe the property, its features, and what makes it special..."
              rows={5}
              {...register("description", {
                required: "Description is required",
                minLength: {
                  value: 50,
                  message: "Description must be at least 50 characters",
                },
              })}
              className={errors.description ? "border-red-500" : ""}
            />
            {errors.description && (
              <p className="text-sm text-red-500">
                {errors.description.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="price">Price *</Label>
            <div className="flex gap-2">
              <Input
                id="price"
                type="number"
                placeholder="e.g., 550000"
                {...register("price", {
                  required: "Price is required",
                  min: { value: 1, message: "Price must be greater than 0" },
                })}
                className={`flex-1 ${errors.price ? "border-red-500" : ""}`}
              />
              <Select
                defaultValue="Dollar"
                onValueChange={(value: Currency) => setValue("currency", value)}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Currency" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Rupees">Rupees</SelectItem>
                  <SelectItem value="Dollar">Dollar</SelectItem>
                  <SelectItem value="Dirhams">Dirhams</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {errors.price && (
              <p className="text-sm text-red-500">{errors.price.message}</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Location Details */}
      <Card>
        <CardHeader>
          <CardTitle>Location Details</CardTitle>
          <CardDescription>Where is this property located?</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="address">Street Address *</Label>
            <Input
              id="address"
              placeholder="e.g., 123 Main Street"
              {...register("address", { required: "Address is required" })}
              className={errors.address ? "border-red-500" : ""}
            />
            {errors.address && (
              <p className="text-sm text-red-500">{errors.address.message}</p>
            )}
          </div>

          <div className="grid md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="city">City *</Label>
              <Input
                id="city"
                placeholder="e.g., Los Angeles"
                {...register("city", { required: "City is required" })}
                className={errors.city ? "border-red-500" : ""}
              />
              {errors.city && (
                <p className="text-sm text-red-500">{errors.city.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="state">State *</Label>
              <Input
                id="state"
                placeholder="e.g., CA"
                {...register("state", { required: "State is required" })}
                className={errors.state ? "border-red-500" : ""}
              />
              {errors.state && (
                <p className="text-sm text-red-500">{errors.state.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="zipCode">ZIP Code *</Label>
              <Input
                id="zipCode"
                placeholder="e.g., 90001 or 845420"
                {...register("zipCode", {
                  required: "ZIP code is required",
                  pattern: {
                    value: /^\d{5,6}(-\d{4})?$/,
                    message: "Invalid ZIP/PIN code format",
                  },
                })}
                className={errors.zipCode ? "border-red-500" : ""}
              />
              {errors.zipCode && (
                <p className="text-sm text-red-500">{errors.zipCode.message}</p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Amenities & Features */}
      <Card>
        <CardHeader>
          <CardTitle>Amenities & Features</CardTitle>
          <CardDescription>
            Add key features and amenities of the property
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Input
              placeholder="e.g., Swimming Pool, Gym, Parking"
              value={amenityInput}
              onChange={(e) => setAmenityInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  addAmenity();
                }
              }}
            />
            <Button
              type="button"
              onClick={() => addAmenity()}
              variant="outline"
            >
              Add
            </Button>
          </div>

          {amenities.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {amenities.map((amenity) => (
                <Badge
                  key={amenity}
                  variant="secondary"
                  className="px-3 py-1 cursor-pointer hover:bg-gray-200"
                >
                  {amenity}
                  <X
                    className="ml-2 h-3 w-3"
                    onClick={() => removeAmenity(amenity)}
                  />
                </Badge>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Image Upload */}
      <Card>
        <CardHeader>
          <CardTitle>Property Images</CardTitle>
          <CardDescription>
            Upload high-quality images of the property
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-indigo-400 transition-colors">
            <input
              type="file"
              id="images"
              accept="image/*"
              multiple
              onChange={(e) => handleImageUpload(e.target.files)}
              className="hidden"
            />
            <label
              htmlFor="images"
              className="cursor-pointer flex flex-col items-center gap-2"
            >
              <Upload className="h-10 w-10 text-gray-400" />
              <div>
                <p className="font-medium text-gray-700">
                  Click to upload images
                </p>
                <p className="text-sm text-gray-500">
                  PNG, JPG, WEBP up to 10MB each
                </p>
              </div>
            </label>
          </div>

          {images.length > 0 && (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {images.map((image, index) => (
                <div key={index} className="relative group">
                  <Image
                    src={image.preview}
                    alt={`Preview ${index + 1}`}
                    width={200}
                    height={128}
                    className="w-full h-32 object-cover rounded-lg"
                  />
                  <button
                    type="button"
                    onClick={() => removeImage(index)}
                    className="absolute top-2 right-2 bg-red-500 text-white rounded-full p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Agent Information */}
      <Card>
        <CardHeader>
          <CardTitle>Agent Information</CardTitle>
          <CardDescription>
            Your contact details for client inquiries
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="agentName">Full Name *</Label>
            <Input
              id="agentName"
              placeholder="e.g., John Smith"
              {...register("agentName", { required: "Agent name is required" })}
              className={errors.agentName ? "border-red-500" : ""}
            />
            {errors.agentName && (
              <p className="text-sm text-red-500">{errors.agentName.message}</p>
            )}
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="agentEmail">Email *</Label>
              <Input
                id="agentEmail"
                type="email"
                placeholder="e.g., john@realestate.com"
                {...register("agentEmail", {
                  required: "Email is required",
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: "Invalid email address",
                  },
                })}
                className={errors.agentEmail ? "border-red-500" : ""}
              />
              {errors.agentEmail && (
                <p className="text-sm text-red-500">
                  {errors.agentEmail.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="agentPhone">Phone Number *</Label>
              <Input
                id="agentPhone"
                type="tel"
                placeholder="e.g., (555) 123-4567"
                {...register("agentPhone", {
                  required: "Phone number is required",
                  pattern: {
                    value: /^[\d\s\-\(\)]+$/,
                    message: "Invalid phone number",
                  },
                })}
                className={errors.agentPhone ? "border-red-500" : ""}
              />
              {errors.agentPhone && (
                <p className="text-sm text-red-500">
                  {errors.agentPhone.message}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Submit Button */}
      <div className="flex gap-4 justify-end">
        {onClose && (
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="bg-indigo-600 hover:bg-indigo-700 min-w-[150px] cursor-pointer"
        >
          {isSubmitting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Submitting...
            </>
          ) : (
            "Submit Listing"
          )}
        </Button>
      </div>
    </form>
  );
}
