import React from "react";
import Header from "@/components/layout/Header";
import HeroMonograph from "@/components/home/HeroMonograph";
import MetricsPlates from "@/components/home/MetricsPlates";
import CoverageDeepDive from "@/components/home/CoverageDeepDive";
import VoicePipelineFlow from "@/components/home/VoicePipelineFlow";
import ResilienceGrid from "@/components/home/ResilienceGrid";
import InterfaceShowcaseTabs from "@/components/home/InterfaceShowcaseTabs";
import TechStackGrid from "@/components/home/TechStackGrid";
import FinalCTA from "@/components/home/FinalCTA";
import Footer from "@/components/layout/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <HeroMonograph />
        <MetricsPlates />
        <CoverageDeepDive />
        <VoicePipelineFlow />
        <ResilienceGrid />
        <InterfaceShowcaseTabs />
        <TechStackGrid />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
