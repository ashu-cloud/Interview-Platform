"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import Logo from "@/components/ui/Logo";
import { api } from "@/lib/api";

export default function AuthPage() {
  const router = useRouter();
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
    if (score < 2) return { score: 1, text: "WEAK", color: "text-[#DC2626]" };
    if (score <= 3) return { score: 2, text: "ROBUST", color: "text-[#D97706]" };
    return { score: 4, text: "ENTROPY SECURE", color: "text-[#16A34A]" };
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
        router.push("/tracks");
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
        router.push("/tracks");
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
    router.push(`/interview/${encodeURIComponent(candidateToken.trim())}`);
  };

  return (
    <div className="min-h-screen bg-[#F5F2EB] text-[#141413] flex flex-col justify-between selection:bg-[#141413] selection:text-[#FAF8F5] relative overflow-hidden font-sans antialiased">
      {/* Background Architectural Drafting Grid & Hairline Axes */}
      <div className="pointer-events-none absolute inset-0 z-0">
        <svg className="w-full h-full stroke-[#141413]/[0.06]" fill="none">
          <defs>
            <pattern id="auth-grid-pattern" width="60" height="60" patternUnits="userSpaceOnUse">
              <path d="M 60 0 L 0 0 0 60" strokeWidth="0.5" />
            </pattern>
            {/* Stipple pattern for left architectural circle */}
            <pattern id="stipple-dense" width="4" height="4" patternUnits="userSpaceOnUse">
              <circle cx="2" cy="2" r="1.1" fill="#141413" fillOpacity="0.5" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#auth-grid-pattern)" />
        </svg>

        {/* Global architectural hairline guides */}
        <div className="absolute top-12 left-0 right-0 h-px bg-[#141413]/[0.08]" />
        <div className="absolute top-0 bottom-0 left-[18%] w-px bg-[#141413]/[0.07] hidden xl:block" />
        <div className="absolute top-0 bottom-0 right-[28%] w-px bg-[#141413]/[0.07] hidden lg:block" />
      </div>

      {/* Left Oversized Stippled Architectural Circle (matching Screen 02 Reference) */}
      <div className="pointer-events-none absolute left-0 top-1/2 -translate-y-1/2 -translate-x-[45%] w-[480px] h-[480px] z-0 hidden lg:block select-none" aria-hidden="true">
        <svg viewBox="0 0 480 480" className="w-full h-full overflow-visible">
          {/* Halftone stippled circle */}
          <circle cx="240" cy="240" r="210" fill="url(#stipple-dense)" />
          {/* Concentric rings */}
          <circle cx="240" cy="240" r="230" stroke="#141413" strokeWidth="0.75" fill="none" opacity="0.3" strokeDasharray="4 4" />
          <circle cx="240" cy="240" r="210" stroke="#141413" strokeWidth="0.8" fill="none" opacity="0.4" />
          <circle cx="240" cy="240" r="140" stroke="#141413" strokeWidth="0.6" fill="none" opacity="0.25" />
          {/* Crosshairs & registration dots */}
          <line x1="240" y1="10" x2="240" y2="470" stroke="#141413" strokeWidth="0.75" opacity="0.3" />
          <line x1="10" y1="240" x2="470" y2="240" stroke="#141413" strokeWidth="0.75" opacity="0.3" />
          {/* Registration coordinate dots */}
          <circle cx="240" cy="30" r="2.5" fill="#141413" opacity="0.7" />
          <circle cx="450" cy="240" r="2.5" fill="#141413" opacity="0.7" />
          <circle cx="360" cy="120" r="2.5" fill="#141413" opacity="0.7" />
          <circle cx="360" cy="360" r="2.5" fill="#141413" opacity="0.7" />
        </svg>
      </div>

      {/* Solid Cobalt Blue Wedge in Bottom Left Corner (matching Screen 02 Reference) */}
      <div className="pointer-events-none absolute bottom-0 left-0 z-10 hidden sm:block">
        <div className="w-24 h-20 bg-[#1D4ED8]" />
      </div>

      {/* Top Bar with Center Title Badge & Back Link matching Screen 02 */}
      <div className="relative z-20 pt-6 px-6 lg:px-12 flex items-center justify-between">
        <div className="flex-1" />
        {/* Pinned Top Badge */}
        <div className="border border-[#D1CCC2] bg-[#FAF8F5] px-4 py-1 rounded-xs shadow-xs">
          <span className="font-mono text-[10px] text-[#767872] uppercase tracking-widest font-semibold">
            VOXASSESS // RECRUITER PORTAL
          </span>
        </div>
        <div className="flex-1 flex justify-end">
          <Link
            href="/"
            className="font-mono text-[11px] uppercase tracking-widest text-[#767872] hover:text-[#141413] transition-colors flex items-center gap-1.5"
          >
            <span>←</span>
            <span>BACK TO HOME</span>
          </Link>
        </div>
      </div>

      {/* Main Studio Work Area */}
      <main className="relative z-10 flex-1 max-w-7xl mx-auto w-full px-6 lg:px-12 py-8 flex items-center justify-center">
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-center">
          {/* Center Column: Authentication Terminal Chassis (7 cols) */}
          <div className="lg:col-span-7 flex flex-col items-center max-w-lg mx-auto w-full">
            {/* Logo Chassis - Restored cleanly without decorative containers */}
            <div className="flex flex-col items-center text-center mb-6">
              <div className="mb-2.5">
                <Logo className="w-16 h-16 sm:w-20 sm:h-20 text-[#141413]" size={72} />
              </div>
              <div className="bg-[#141413] text-[#FAF8F5] px-2.5 py-0.5 rounded-xs mb-3 font-mono text-[9px] uppercase tracking-widest font-semibold shadow-xs">
                VOXI // SENTINEL AUTH
              </div>

              <div className="flex items-center gap-2 mb-1.5">
                <span className="font-mono text-[10px] font-bold text-[#1D4ED8] uppercase tracking-widest">
                  AUTHENTICATION PORTAL
                </span>
                <span className="text-[#C7C7C0]">/</span>
                <span className="font-mono text-[10px] text-[#767872] uppercase tracking-wider">
                  SECURE ACCESS
                </span>
              </div>

              <h1 className="font-serif text-3xl font-normal text-[#141413] tracking-tight">
                {activeTab === "signin" ? "Sign In to VoxAssess" : "Register New Tenant"}
              </h1>
              <p className="font-sans text-xs text-[#767872] max-w-sm mt-1.5 leading-relaxed">
                Enterprise tenant management and hiring calibration for the autonomous voice technical assessment platform.
              </p>
            </div>

            {/* Tab Switcher: Sign In vs Register matching Screen 02 */}
            <div className="w-full bg-[#E8E4DB] p-1 mb-5 grid grid-cols-2 gap-1 border border-[#D1CCC2] rounded">
              <button
                type="button"
                onClick={() => setActiveTab("signin")}
                className={`py-2 px-3 font-mono text-[11px] uppercase tracking-wider transition-all rounded-xs font-semibold ${
                  activeTab === "signin"
                    ? "bg-[#FAF8F5] text-[#141413] shadow-xs border-b-2 border-[#141413]"
                    : "text-[#767872] hover:text-[#141413] bg-transparent"
                }`}
              >
                SIGN IN // CONSOLE
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("register")}
                className={`py-2 px-3 font-mono text-[11px] uppercase tracking-wider transition-all rounded-xs font-semibold ${
                  activeTab === "register"
                    ? "bg-[#FAF8F5] text-[#141413] shadow-xs border-b-2 border-[#141413]"
                    : "text-[#767872] hover:text-[#141413] bg-transparent"
                }`}
              >
                REGISTER // NEW TENANT
              </button>
            </div>

            {errorMessage && (
              <div className="w-full mb-4 p-3 bg-[#FEF2F2] border border-[#EF4444] rounded text-xs font-mono text-[#991B1B] leading-relaxed animate-in fade-in">
                ⚠ {errorMessage}
              </div>
            )}

            {authStatus && (
              <div className="w-full mb-4 p-3 bg-[#F0FDF4] border border-[#16A34A] rounded text-xs font-mono text-[#166534] leading-relaxed animate-in fade-in">
                {authStatus}
              </div>
            )}

            {/* Form A: Sign In */}
            {activeTab === "signin" && (
              <form onSubmit={handleSignIn} className="w-full flex flex-col gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                    WORK EMAIL / PRINCIPAL
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="akky@ocnt.com"
                    value={signInEmail}
                    onChange={(e) => setSignInEmail(e.target.value)}
                    className="px-3.5 py-2.5 text-xs font-mono bg-[#EBF0FF]/70 border border-[#CBD5E1] rounded focus:border-[#1D4ED8] focus:bg-white focus:outline-none transition-colors text-[#141413]"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <div className="flex items-center justify-between">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      PASSWORD / CREDENTIAL
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowSignInPassword(!showSignInPassword)}
                      className="font-mono text-[9px] text-[#767872] hover:text-[#141413] uppercase font-bold"
                    >
                      {showSignInPassword ? "[ CONCEAL KEY ]" : "[ REVEAL KEY ]"}
                    </button>
                  </div>
                  <input
                    type={showSignInPassword ? "text" : "password"}
                    required
                    placeholder="••••••••••••"
                    value={signInPassword}
                    onChange={(e) => setSignInPassword(e.target.value)}
                    className="px-3.5 py-2.5 text-xs font-mono bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full mt-2 py-3 bg-[#141413] hover:bg-[#2A2A28] text-white rounded font-mono text-xs uppercase tracking-widest font-semibold transition-colors flex items-center justify-center gap-2 shadow-xs cursor-pointer disabled:opacity-50"
                >
                  <span>{loading ? "AUTHENTICATING..." : "AUTHENTICATE & ENTER STUDIO"}</span>
                  <span className="text-sm">→</span>
                </button>
              </form>
            )}

            {/* Form B: Register Organization */}
            {activeTab === "register" && (
              <form onSubmit={handleRegister} className="w-full flex flex-col gap-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      ORGANIZATION NAME
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="Acme Systems"
                      value={regOrgName}
                      onChange={(e) => handleOrgNameChange(e.target.value)}
                      className="px-3.5 py-2 text-xs font-sans bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      WORKSPACE SLUG
                    </label>
                    <input
                      type="text"
                      required
                      value={regOrgSlug}
                      onChange={(e) => setRegOrgSlug(e.target.value)}
                      className="px-3.5 py-2 text-xs font-mono bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      ADMINISTRATOR NAME
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="Elena Rostova"
                      value={regFullName}
                      onChange={(e) => setRegFullName(e.target.value)}
                      className="px-3.5 py-2 text-xs font-sans bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      WORK EMAIL
                    </label>
                    <input
                      type="email"
                      required
                      placeholder="elena@acme.com"
                      value={regEmail}
                      onChange={(e) => setRegEmail(e.target.value)}
                      className="px-3.5 py-2 text-xs font-mono bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                    />
                  </div>
                </div>

                <div className="flex flex-col gap-1.5">
                  <div className="flex items-center justify-between">
                    <label className="font-mono text-[10px] text-[#767872] uppercase tracking-wider font-semibold">
                      PASSWORD (MIN 8 CHARS)
                    </label>
                    <div className="flex items-center gap-2">
                      <span className={`font-mono text-[9px] font-semibold ${strength.color}`}>
                        {strength.text}
                      </span>
                      <button
                        type="button"
                        onClick={() => setShowRegPassword(!showRegPassword)}
                        className="font-mono text-[9px] text-[#767872] hover:text-[#141413] uppercase font-bold"
                      >
                        {showRegPassword ? "[ CONCEAL ]" : "[ REVEAL ]"}
                      </button>
                    </div>
                  </div>
                  <input
                    type={showRegPassword ? "text" : "password"}
                    required
                    placeholder="••••••••••••"
                    value={regPassword}
                    onChange={(e) => setRegPassword(e.target.value)}
                    className="px-3.5 py-2.5 text-xs font-mono bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                  />

                  {/* Password Strength Meter */}
                  <div className="grid grid-cols-4 gap-1.5 mt-1">
                    {[1, 2, 3, 4].map((step) => (
                      <div
                        key={step}
                        className={`h-1 rounded-full transition-all ${
                          strength.score >= step
                            ? strength.score >= 3
                              ? "bg-[#16A34A]"
                              : "bg-[#D97706]"
                            : "bg-[#D1CCC2]"
                        }`}
                      />
                    ))}
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full mt-2 py-3 bg-[#141413] hover:bg-[#2A2A28] text-white rounded font-mono text-xs uppercase tracking-widest font-semibold transition-colors flex items-center justify-center gap-2 shadow-xs cursor-pointer disabled:opacity-50"
                >
                  <span>{loading ? "PROVISIONING..." : "PROVISION ORGANIZATION TENANT"}</span>
                  <span className="text-sm">→</span>
                </button>
              </form>
            )}
          </div>

          {/* Right Column: Architectural Drafting Graphic & Candidate Entry Box (5 cols) matching Screen 02 */}
          <div className="lg:col-span-5 flex flex-col items-center lg:items-start gap-8">
            {/* Signature Circular Blue-and-Grid Drafting Diagram matching Screen 02 Reference */}
            <div className="relative flex items-center gap-6">
              <svg width="250" height="250" viewBox="0 0 250 250" className="overflow-visible select-none" aria-hidden="true">
                <defs>
                  <pattern id="auth-stipple-dense" width="4" height="4" patternUnits="userSpaceOnUse">
                    <circle cx="2" cy="2" r="1.1" fill="#141413" fillOpacity="0.5" />
                  </pattern>
                </defs>

                {/* Outer concentric hairline circles */}
                <circle cx="125" cy="125" r="115" stroke="#141413" strokeWidth="0.8" strokeDasharray="3 3" fill="none" opacity="0.3" />
                <circle cx="125" cy="125" r="100" stroke="#141413" strokeWidth="0.8" fill="none" opacity="0.4" />
                <circle cx="125" cy="125" r="70" stroke="#141413" strokeWidth="0.6" fill="none" opacity="0.25" />

                {/* Ultramarine / Cobalt blue semi-circle on the left */}
                <path
                  d="M 125 25 A 100 100 0 0 0 125 225 Z"
                  fill="#1D4ED8"
                />

                {/* Top-right quadrant with stippling texture */}
                <path
                  d="M 125 25 A 100 100 0 0 1 225 125 L 125 125 Z"
                  fill="url(#auth-stipple-dense)"
                />

                {/* Horizontal & vertical axis crosshairs */}
                <line x1="5" y1="125" x2="245" y2="125" stroke="#141413" strokeWidth="0.8" opacity="0.4" />
                <line x1="125" y1="5" x2="125" y2="245" stroke="#141413" strokeWidth="0.8" opacity="0.4" />

                {/* Top registration tick with letter 'A' */}
                <line x1="125" y1="5" x2="125" y2="15" stroke="#141413" strokeWidth="1.2" />
                <text x="125" y="0" textAnchor="middle" className="font-mono text-[9px] fill-[#141413] font-bold">A</text>

                {/* Center marker */}
                <circle cx="125" cy="125" r="3.5" fill="#141413" />
                <circle cx="125" cy="125" r="1.5" fill="#FAF8F5" />
              </svg>

              {/* Stacked technical metadata labels */}
              <div className="flex flex-col gap-2 font-mono text-[11px] uppercase tracking-widest text-[#767872]">
                <span className="text-[#141413] font-bold">SECURE</span>
                <span className="text-[#141413]/70 font-medium">EVALUATE</span>
                <span className="text-[#767872]">SCALE</span>
              </div>
            </div>

            {/* Candidate / Participant Fast Entry Card */}
            <div className="w-full max-w-sm bg-[#FAF8F5] border border-[#D1CCC2] p-5 rounded relative shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="font-mono text-[10px] uppercase tracking-wider font-bold text-[#141413]">
                  CANDIDATE OR PARTICIPANT ENTRY
                </span>
                <span className="font-mono text-[9px] text-[#767872] uppercase tracking-wider font-medium">
                  1-USE ACCREDITED TOKEN
                </span>
              </div>

              <p className="font-sans text-xs text-[#767872] leading-relaxed mb-3">
                Candidates never require passwords or accounts. Check your interview email for your unique single-use token (interview-token).
              </p>

              <form onSubmit={handleCandidateToken} className="flex gap-2">
                <input
                  type="text"
                  placeholder="enter stage token (e.g. 9104-74a0...)"
                  value={candidateToken}
                  onChange={(e) => setCandidateToken(e.target.value)}
                  className="flex-1 px-3 py-1.5 text-xs font-mono bg-white border border-[#D1CCC2] rounded focus:border-[#141413] focus:outline-none transition-colors"
                />
                <button
                  type="submit"
                  className="px-3.5 py-1.5 bg-[#141413] hover:bg-[#2A2A28] text-white transition-colors text-[11px] font-mono uppercase tracking-wider rounded font-medium shadow-xs cursor-pointer whitespace-nowrap"
                >
                  ENTER STAGE
                </button>
              </form>
            </div>
          </div>
        </div>
      </main>

      {/* Footer Hairline Notice */}
      <footer className="relative z-10 border-t border-[#D1CCC2] bg-[#F5F2EB] py-3">
        <div className="max-w-7xl mx-auto px-6 lg:px-12 flex flex-col sm:flex-row items-center justify-between gap-2 text-[10px] font-mono text-[#767872] uppercase tracking-wider">
          <div>VOXASSESS SYSTEM ENGINE // ARCHIVAL INSTRUMENT v4.2</div>
          <div>DETERMINISTIC EVALUATION MESH // AES-256 ZERO RETENTION</div>
        </div>
      </footer>
    </div>
  );
}
