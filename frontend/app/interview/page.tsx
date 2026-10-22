"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Logo from "@/components/ui/Logo";
import { api } from "@/lib/api";

export default function InterviewAccessPage() {
  const router = useRouter();
  const [tokenInput, setTokenInput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    setIsLoggedIn(api.isAuthenticated());
  }, []);

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    const cleanToken = tokenInput.trim();
    if (!cleanToken) {
      setError("Please paste a valid assessment invitation token.");
      return;
    }
    // Clean URL if full invite link was pasted
    const extractedToken = cleanToken.includes("/interview/")
      ? cleanToken.split("/interview/")[1]?.split("?")[0]
      : cleanToken;

    if (!extractedToken) {
      setError("Unable to parse token from the provided link.");
      return;
    }

    router.push(`/interview/${extractedToken}`);
  };

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex flex-col justify-between text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <header className="h-16 px-6 lg:px-12 flex items-center justify-between border-b border-[#e5e2da]">
        <Link href="/" className="flex items-center gap-2.5">
          <Logo className="w-5 h-5 text-sumi" />
          <span className="font-mono text-xs uppercase tracking-widest font-bold text-sumi">
            VoxAssess
          </span>
        </Link>
        <Link
          href={isLoggedIn ? "/console" : "/auth"}
          className="text-xs font-mono text-sumi-secondary hover:text-sumi transition-colors flex items-center gap-1.5"
        >
          {isLoggedIn && <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />}
          <span>{isLoggedIn ? "Open Console →" : "Hiring Team Console →"}</span>
        </Link>
      </header>

      <main className="flex-1 flex items-center justify-center p-6">
        <div className="max-w-md w-full bg-white border border-[#e5e2da] rounded-xs shadow-xs p-8 sm:p-10">
          <div className="flex items-center gap-2 font-mono text-[11px] text-sumi-muted uppercase tracking-wider mb-2">
            <span>Stage Gateway</span>
            <span className="text-[#c7c7c0]">/</span>
            <span className="text-emerald-700 font-semibold">Token Verification</span>
          </div>

          <h1 className="font-sans text-2xl font-medium text-sumi tracking-tight mb-2">
            Enter Interview Stage
          </h1>
          <p className="font-sans text-xs text-sumi-secondary leading-relaxed mb-6">
            Please enter your personalized candidate invitation token or paste your invitation link to commence the technical assessment.
          </p>

          <form onSubmit={handleJoin} className="space-y-4">
            <div>
              <label className="block font-mono text-[11px] uppercase tracking-wider text-sumi mb-1.5 font-medium">
                Invitation Token or Link
              </label>
              <input
                type="text"
                value={tokenInput}
                onChange={(e) => {
                  setTokenInput(e.target.value);
                  if (error) setError(null);
                }}
                placeholder="Paste token or link (e.g. 47b978eb-8131...)"
                className="w-full px-3.5 py-2.5 bg-[#fbf9f5] border border-[#d5d2c8] focus:border-sumi focus:outline-none rounded-xs font-mono text-xs text-sumi placeholder:text-neutral-400"
              />
            </div>

            {error && (
              <div className="p-3 bg-red-50 border border-red-200 text-red-700 font-mono text-xs rounded-xs">
                {error}
              </div>
            )}

            <button
              type="submit"
              className="w-full py-2.5 bg-sumi text-white hover:bg-neutral-800 transition-colors font-mono text-xs font-medium rounded-xs shadow-xs cursor-pointer"
            >
              Enter Stage →
            </button>
          </form>

          <div className="mt-8 pt-6 border-t border-[#e5e2da] flex items-center justify-between text-xs font-mono text-sumi-muted">
            <span>Audio &amp; Video Ready</span>
            <span className="text-emerald-600">Runtime Ready</span>
          </div>
        </div>
      </main>

      <footer className="h-12 px-6 flex items-center justify-center border-t border-[#e5e2da] text-xs font-mono text-sumi-muted">
        <span>Autonomous Technical Assessment Infrastructure</span>
      </footer>
    </div>
  );
}
