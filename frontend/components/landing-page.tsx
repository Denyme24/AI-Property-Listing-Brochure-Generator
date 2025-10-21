"use client";

import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/navbar";
import { Sparkles } from "lucide-react";

interface LandingPageProps {
  onGetStarted: () => void;
}

export function LandingPage({ onGetStarted }: LandingPageProps) {
  return (
    <div className="h-screen w-full relative overflow-hidden bg-linear-to-br from-[#002349] via-[#003d7a] to-[#002349]">
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute top-1/4 -left-1/4 w-96 h-96 bg-[#4A8BDF] rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob"></div>
        <div className="absolute top-1/3 -right-1/4 w-96 h-96 bg-[#178582] rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000"></div>
        <div className="absolute -bottom-1/4 left-1/2 w-96 h-96 bg-[#BFA181] rounded-full mix-blend-multiply filter blur-3xl opacity-10 animate-blob animation-delay-4000"></div>
      </div>

      {/* Grid pattern overlay */}
      <div className="absolute inset-0  opacity-30"></div>

      {/* Navbar */}
      <Navbar onGetStarted={onGetStarted} />

      {/* Hero Section */}
      <section className="relative h-full flex items-center justify-center">
        <div className="max-w-5xl mx-auto text-center space-y-6 px-4 animate-fade-in">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-5 py-2 bg-white/10 backdrop-blur-sm border border-white/20 text-white rounded-full text-sm font-medium shadow-lg animate-slide-down">
            <Sparkles className="h-4 w-4 text-[#BFA181] animate-pulse" />
            AI-Powered Property Marketing
          </div>

          {/* Main Headline */}
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold text-white leading-tight animate-slide-up">
            Create Stunning{" "}
            <span className="bg-linear-to-r from-[#4A8BDF] via-[#BFA181] to-[#178582] bg-clip-text text-transparent animate-gradient">
              Property Brochures
            </span>{" "}
            in Minutes
          </h1>

          {/* Subtitle */}
          <p className="text-lg md:text-xl text-gray-300 max-w-3xl mx-auto leading-relaxed animate-slide-up animation-delay-200">
            Transform your property listings into professional, AI-generated
            brochures that captivate buyers and close deals faster.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4 animate-slide-up animation-delay-400">
            <Button
              onClick={onGetStarted}
              size="lg"
              className="bg-[#BFA181] hover:bg-[#957C3D] text-white text-base px-8 py-6 shadow-2xl hover:shadow-[#BFA181]/50 transition-all duration-300 transform hover:scale-105 border-0 font-semibold cursor-pointer"
            >
              Get Started Free
              <span className="ml-2 transition-transform group-hover:translate-x-1">
                →
              </span>
            </Button>
            <Button
              variant="outline"
              size="lg"
              className="bg-transparent border-2 border-white text-white hover:bg-white hover:text-[#002349] text-base px-8 py-6 transition-all duration-300 transform hover:scale-105 font-semibold cursor-pointer"
            >
              Watch Demo
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}
