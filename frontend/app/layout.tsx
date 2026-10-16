import type { Metadata } from "next";
import { Space_Grotesk, Cormorant_Garamond, JetBrains_Mono } from "next/font/google";
import CustomCursor from "@/components/ui/CustomCursor";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  variable: "--font-space-grotesk",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  display: "swap",
});

const cormorant = Cormorant_Garamond({
  variable: "--font-cormorant",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600"],
  style: ["normal", "italic"],
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "VoxAssess // Autonomous Technical Voice Assessment Platform",
  description:
    "Sub-800ms bidirectional conversational voice powered by sentence-boundary streaming, mathematical cosine centroid coverage tracking, and tamper-proof verbatim substring-verified transcript attribution.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${spaceGrotesk.variable} ${cormorant.variable} ${jetbrainsMono.variable} scroll-smooth`}
    >
      <body
        suppressHydrationWarning
        className="min-h-screen technical-grid selection:bg-sumi selection:text-paper antialiased flex flex-col justify-between"
      >
        <CustomCursor />
        {children}
      </body>
    </html>
  );
}
