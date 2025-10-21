"use client";

import { useState } from "react";
import { LandingPage } from "@/components/landing-page";
import { PropertyForm } from "@/components/property-form";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export default function Home() {
  const [isFormOpen, setIsFormOpen] = useState(false);

  return (
    <>
      <LandingPage onGetStarted={() => setIsFormOpen(true)} />

      <Dialog open={isFormOpen} onOpenChange={setIsFormOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-2xl">
              Submit Your Property Listing
            </DialogTitle>
          </DialogHeader>
          <PropertyForm onClose={() => setIsFormOpen(false)} />
        </DialogContent>
      </Dialog>
    </>
  );
}
