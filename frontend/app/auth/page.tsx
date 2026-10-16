"use client";

import React, { useState } from "react";
import Link from "next/link";
import Logo from "@/components/ui/Logo";
import { api } from "@/lib/api";

export default function AuthPage() {
  const [activeTab, setActiveTab] = useState<"signin" | "register">("signin");
  const [showSignInPassword, setShowSignInPassword] = useState(false);
  const [showRegPassword, setShowRegPassword] = useState(false);

  // Form states
  const [signInEmail, setSignInEmail] = useState("");
  const [signInPassword, setSignInPassword] = useState("");

  const [regOrgName, setRegOrgName] = useState("");
  const [regOrgSlug, setRegOrgSlug] = useState("");
  const [regFullName, setRegFullName] = useState("");
  const [regEmail, setRegEmail] = useState("");
  const [regPassword, setRegPassword] = useState("");

  const [candidateToken, setCandidateToken] = useState("");
  const [authStatus, setAuthStatus] = useState<string | null>(null);

  // Slug auto-generation
  const handleOrgNameChange = (val: string) => {
    setRegOrgName(val);
    const slug = val
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^\w-]+/g, "")
      .replace(/--+/g, "-")
      .replace(/^-+/, "")
      .replace(/-+$/, "");
    setRegOrgSlug(slug || "acme-systems");
  };

  // Password entropy calculation
  const getPasswordStrength = (pass: string) => {
    let score = 0;
    if (pass.length >= 8) score++;
    if (/[A-Z]/.test(pass) && /[a-z]/.test(pass)) score++;
    if (/[0-9]/.test(pass)) score++;
    if (/[^A-Za-z0-9]/.test(pass)) score++;

    if (pass.length === 0) return { score: 0, text: "MIN 8 CH", color: "text-[#767872]" };
    if (score < 2) return { score: 1, text: "WEAK", color: "text-vermilion" };
    if (score <= 3) return { score: 2, text: "ROBUST", color: "text-amber-750" };
    return { score: 4, text: "ENTROPY SECURE", color: "text-emerald-600" };
  };

  const strength = getPasswordStrength(regPassword);

  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleSignIn = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setLoading(true);
    setAuthStatus("Authenticating against platform-service...");

    try {
      const res = await api.login({
        email: signInEmail.trim(),
        password: signInPassword,
      });
      setAuthStatus(
        `[200 OK] Handshake authenticated for ${res.user?.email || signInEmail}. Bearer JWT issued. Redirecting...`
      );
      setTimeout(() => {
        window.location.href = "/tracks";
      }, 800);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Authentication failed";
      setErrorMessage(msg);
      setAuthStatus(null);
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setLoading(true);
    setAuthStatus("Provisioning tenant organization and author account...");

    try {
      const res = await api.register({
        email: regEmail.trim(),
        password: regPassword,
        full_name: regFullName.trim(),
        org_name: regOrgName.trim(),
        org_slug: regOrgSlug.trim(),
      });
      setAuthStatus(
        `[201 CREATED] Tenant organization '${res.org?.name || regOrgName}' initialized. Redirecting...`
      );
      setTimeout(() => {
        window.location.href = "/tracks";
      }, 800);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Tenant registration failed";
      setErrorMessage(msg);
      setAuthStatus(null);
    } finally {
      setLoading(false);
    }
  };

  const handleCandidateToken = (e: React.FormEvent) => {
    e.preventDefault();
    if (!candidateToken.trim()) return;
    window.location.href = `/interview/${encodeURIComponent(candidateToken.trim())}`;
  };

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex flex-col justify-between text-sumi selection:bg-sumi selection:text-paper antialiased">
      {/* Top Navigation */}
      <header className="sticky top-0 z-40 bg-[#fbf9f5]/90 backdrop-blur-md border-b border-[#e5e2da]">
        <div className="h-16 max-w-7xl mx-auto px-6 lg:px-12 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/" className="flex items-center gap-3 group" title="Home">
              <Logo className="w-8 h-8 text-sumi group-hover:scale-105 transition-transform" size={32} />
            </Link>
          </div>

          <nav className="flex items-center gap-4 text-xs font-mono uppercase tracking-widest text-sumi-secondary">
            <Link href="/" className="hover:text-sumi transition-colors">
              Back to Home
            </Link>
          </nav>
        </div>
      </header>

      {/* Centered Auth Folio Chassis */}
      <main className="flex-1 flex flex-col items-center justify-center px-4 sm:px-6 py-12 relative">
        {/* Subtle Archival Drafting Crosshairs */}
        <div className="w-full max-w-xl relative bg-[#f5f3ef] border border-[#e5e2da] p-6 sm:p-10 rounded shadow-[0_12px_40px_-16px_rgba(20,20,19,0.06)]">
          {/* Security Gate Badge */}
          <div className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-[#fbf9f5] px-3 py-0.5 border border-[#e5e2da] rounded-xs shadow-xs">
            <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-widest font-semibold">
              VOXASSESS // RECRUITER PORTAL
            </span>
          </div>

          {/* Header & Logo with Dynamic Eye Tracking */}
          <div className="flex flex-col items-center text-center mb-8">
            <div className="relative mb-6 p-5 sm:p-6 bg-[#fbf9f5] border border-[#e5e2da] rounded shadow-plate group">
              <Logo className="w-32 h-32 sm:w-36 sm:h-36 text-sumi" size={144} />
              <div className="absolute -bottom-2.5 left-1/2 -translate-x-1/2 bg-sumi px-2.5 py-0.5 whitespace-nowrap shadow-xs rounded-xs">
                <span className="font-mono text-[9px] uppercase tracking-widest text-[#fbf9f5] font-semibold">
                  VOXI // SENTINEL AUTH
                </span>
              </div>
            </div>

            <div className="flex items-center gap-2 mb-1">
              <span className="font-mono text-[11px] font-bold text-ultramarine uppercase tracking-widest">
                AUTHENTICATION PORTAL
              </span>
              <span className="text-[#c7c7c0]">/</span>
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider">
                SECURE ACCESS
              </span>
            </div>

            <h1 className="font-sans text-2xl sm:text-3xl font-medium text-sumi tracking-tight">
              Sign In to VoxAssess
            </h1>
            <p className="font-sans text-xs sm:text-sm text-sumi-secondary max-w-md mt-1.5 leading-relaxed">
              Enterprise tenant management and hiring calibration for the autonomous voice technical assessment platform.
            </p>
          </div>

          {/* Tab Switcher: Sign In vs Register */}
          <div className="w-full bg-[#efeeea] p-1 mb-6 grid grid-cols-2 gap-1 border border-[#e5e2da] rounded">
            <button
              type="button"
              onClick={() => setActiveTab("signin")}
              className={`py-2 px-3 font-mono text-xs uppercase tracking-wider transition-all rounded-xs font-semibold ${
                activeTab === "signin"
                  ? "bg-[#fbf9f5] text-sumi shadow-xs border-b-2 border-sumi"
                  : "text-sumi-muted hover:text-sumi bg-transparent"
              }`}
            >
              Sign In // Console
            </button>
            <button
              type="button"
              onClick={() => setActiveTab("register")}
              className={`py-2 px-3 font-mono text-xs uppercase tracking-wider transition-all rounded-xs font-semibold ${
                activeTab === "register"
                  ? "bg-[#fbf9f5] text-sumi shadow-xs border-b-2 border-sumi"
                  : "text-sumi-muted hover:text-sumi bg-transparent"
              }`}
            >
              Register // New Tenant
            </button>
          </div>

          {errorMessage && (
            <div className="mb-6 p-3 bg-red-50 border border-red-500 rounded text-xs font-mono text-red-700 leading-relaxed animate-in fade-in">
              ⚠ {errorMessage}
            </div>
          )}

          {authStatus && (
            <div className="mb-6 p-3 bg-white border border-emerald-600 rounded text-xs font-mono text-emerald-800 leading-relaxed animate-in fade-in">
              {authStatus}
            </div>
          )}

          {/* Form A: Sign In */}
          {activeTab === "signin" && (
            <form onSubmit={handleSignIn} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                  Work Email / Principal
                </label>
                <input
                  type="email"
                  required
                  placeholder="systemslead@infrastructure.ai"
                  value={signInEmail}
                  onChange={(e) => setSignInEmail(e.target.value)}
                  className="px-3.5 py-2.5 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Passphrase / Credential
                  </label>
                  <button
                    type="button"
                    onClick={() => setShowSignInPassword(!showSignInPassword)}
                    className="font-mono text-[10px] text-sumi-muted hover:text-sumi uppercase"
                  >
                    {showSignInPassword ? "[ CONCEAL KEY ]" : "[ REVEAL KEY ]"}
                  </button>
                </div>
                <input
                  type={showSignInPassword ? "text" : "password"}
                  required
                  placeholder="••••••••••••••••"
                  value={signInPassword}
                  onChange={(e) => setSignInPassword(e.target.value)}
                  className="px-3.5 py-2.5 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                />
              </div>

              <button
                type="submit"
                className="w-full mt-2 py-3 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs uppercase tracking-widest font-semibold transition-colors flex items-center justify-center gap-2 shadow-sm"
              >
                <span>Authenticate &amp; Enter Studio</span>
                <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </form>
          )}

          {/* Form B: Register Organization */}
          {activeTab === "register" && (
            <form onSubmit={handleRegister} className="flex flex-col gap-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Organization Name
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Acme Systems"
                    value={regOrgName}
                    onChange={(e) => handleOrgNameChange(e.target.value)}
                    className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Workspace Slug
                  </label>
                  <input
                    type="text"
                    required
                    value={regOrgSlug}
                    onChange={(e) => setRegOrgSlug(e.target.value)}
                    className="px-3.5 py-2 text-sm font-mono bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Administrator Name
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Elena Rostova"
                    value={regFullName}
                    onChange={(e) => setRegFullName(e.target.value)}
                    className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Work Email
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="elena@acme.com"
                    value={regEmail}
                    onChange={(e) => setRegEmail(e.target.value)}
                    className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                  />
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between">
                  <label className="font-mono text-[11px] text-sumi-muted uppercase tracking-wider font-semibold">
                    Password (Min 8 chars)
                  </label>
                  <div className="flex items-center gap-2">
                    <span className={`font-mono text-[10px] font-semibold ${strength.color}`}>
                      {strength.text}
                    </span>
                    <button
                      type="button"
                      onClick={() => setShowRegPassword(!showRegPassword)}
                      className="font-mono text-[10px] text-sumi-muted hover:text-sumi uppercase"
                    >
                      {showRegPassword ? "[ CONCEAL ]" : "[ REVEAL ]"}
                    </button>
                  </div>
                </div>
                <input
                  type={showRegPassword ? "text" : "password"}
                  required
                  placeholder="••••••••••••••••"
                  value={regPassword}
                  onChange={(e) => setRegPassword(e.target.value)}
                  className="px-3.5 py-2.5 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
                />

                {/* Password Strength Meter */}
                <div className="grid grid-cols-4 gap-1.5 mt-1">
                  {[1, 2, 3, 4].map((step) => (
                    <div
                      key={step}
                      className={`h-1 rounded-full transition-all ${
                        strength.score >= step
                          ? strength.score >= 3
                            ? "bg-emerald-500"
                            : "bg-amber-750"
                          : "bg-[#e5e2da]"
                      }`}
                    ></div>
                  ))}
                </div>
              </div>

              <button
                type="submit"
                className="w-full mt-2 py-3 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs uppercase tracking-widest font-semibold transition-colors flex items-center justify-center gap-2 shadow-sm"
              >
                <span>Provision Organization Tenant</span>
                <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </form>
          )}

          {/* Participant Invite Fast Jump */}
          <div className="mt-8 pt-6 border-t border-[#e5e2da] flex flex-col gap-2.5">
            <div className="flex items-center justify-between">
              <span className="font-mono text-[10px] uppercase tracking-wider font-bold text-sumi-muted">
                Candidate or Participant Entry
              </span>
              <span className="font-mono text-[10px] text-sumi-muted">1-USE ACCREDITED TOKEN</span>
            </div>
            <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
              Candidates never require passwords or accounts. Check your interview email for your unique single-use token (`/interview/:token`).
            </p>
            <form onSubmit={handleCandidateToken} className="flex gap-2 mt-1">
              <input
                type="text"
                placeholder="Paste token (e.g. tok_7f6d19a4...)"
                value={candidateToken}
                onChange={(e) => setCandidateToken(e.target.value)}
                className="flex-1 px-3 py-1.5 text-xs font-mono bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
              />
              <button
                type="submit"
                className="px-3.5 py-1.5 bg-[#eae8e4] hover:bg-sumi hover:text-white transition-colors text-xs font-mono uppercase tracking-wider rounded font-medium"
              >
                Enter Stage
              </button>
            </form>
          </div>
        </div>
      </main>
    </div>
  );
}
