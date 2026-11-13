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
import {
  ChapterBridge,
  DraftingCursorReticle,
  EditorialMotionProvider,
  ScrollFolioRail,
} from "@/components/motion/ScrollChoreography";

export default function Home() {
  return (
    <EditorialMotionProvider>
      <Header />
      <DraftingCursorReticle />
      <ScrollFolioRail />
      <main className="bg-[#FBF9F5] overflow-x-clip">
        <HeroMonograph />
        <ChapterBridge title="our benchmarks" folioTag="FOLIO 01 → 02 // VERIFIED TELEMETRY" />
        <MetricsPlates />
        <ChapterBridge title="adaptive coverage" folioTag="FOLIO 02 → 03 // CENTROID MANIFOLD" />
        <CoverageDeepDive />
        <ChapterBridge title="our workflow" folioTag="FOLIO 03 → 04 // VOICE PIPELINE" />
        <VoicePipelineFlow />
        <ChapterBridge title="resilience topology" folioTag="FOLIO 04 → 05 // DUAL-TRACK ARCHITECTURE" />
        <ResilienceGrid />
        <ChapterBridge title="interface studio" folioTag="FOLIO 05 → 06 // OPERATOR CONSOLES" />
        <InterfaceShowcaseTabs />
        <ChapterBridge title="scatter-gather mesh" folioTag="FOLIO 06 → 07 // PROVENANCE MATRIX" />
        <TechStackGrid />
        <ChapterBridge title="production stage" folioTag="FOLIO 07 → 08 // SELF-HOST DEPLOYMENT" />
        <FinalCTA />
      </main>
      <Footer />
    </EditorialMotionProvider>
  );
}
