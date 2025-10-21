"use client";

import { Button } from "@/components/ui/button";

interface NavbarProps {
  onGetStarted: () => void;
}

export function Navbar({ onGetStarted }: NavbarProps) {
  return (
    <nav className="absolute top-0 left-0 right-0 z-50">
      <div className="container mx-auto px-6 py-6">
        <div className="flex items-center justify-between">
          {/* Logo/Brand */}
          <div className="flex items-center">
            <span className="text-2xl font-bold text-white">
              PropBrochure AI
            </span>
          </div>

          {/* CTA Button */}
          <Button
            onClick={onGetStarted}
            className="bg-[#BFA181] hover:bg-[#957C3D] text-white shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-105 border-0 cursor-pointer"
            size="lg"
          >
            List Your Property
          </Button>
        </div>
      </div>
    </nav>
  );
}
