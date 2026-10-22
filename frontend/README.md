# VoxAssess — Web Application Frontend

The frontend interface for VoxAssess, built with **Next.js 16 (App Router)**, **Turbopack**, **TypeScript**, and **Tailwind CSS**. It provides a real-time conversational voice interview portal for candidates and an assessment management hub for hiring managers.

---

## Key Features & Capabilities

### 1. Real-Time Candidate Assessment Stage (`/interview/[token]`)
- **Web Audio Ingestion**: Captures linear 16kHz 16-bit PCM mono audio using standard Web Audio API and `ScriptProcessorNode` / `AudioWorklet` buffers.
- **Anti-Aliasing Filter & Pre-Roll Ring Buffer**:
  - Implements an internal linear box filter to downsample from native 44.1/48kHz to 16kHz without Nyquist foldover distortion.
  - Maintains a ~250ms circular audio buffer flushed on speech detection to preserve opening syllables from being clipped.
- **Full-Duplex WebSocket Streaming**: Streams binary Tag `0x01` PCM frames to `interview-orchestrator` while receiving binary Tag `0x02` 24kHz MP3 chunks and JSON control signals.
- **Instant Barge-In Interruption**: Halts HTML5 `AudioContext` playback on user speech detection, clears local audio queues, and synchronizes atomic turn counters with the backend.
- **Diagnostics Overlay HUD (`?debug=1`)**: Appending `?debug=1` to any interview URL reveals a live telemetry HUD showing connection health, audio queue depth, and round-trip turn latency.

### 2. Public Candidate Self-Application (`/apply/:trackId`)
- **Unauthenticated Application**: Candidates apply with their full name and email without requiring administrative login credentials.
- **Instant Stage Entry**: Fetches public track and blueprint details, generates a single-use token, and automatically redirects the candidate into their personalized interview stage.

### 3. Hiring Manager Console (`/console`)
- **Canonical Dashboard (`/console`)**: Real-time overview of active candidate pipelines, completed sessions, and blueprint coverage.
- **Blueprint Studio (`/blueprints/new`)**: Visual authoring tool for technical questions, follow-up hints, and weighted rubric criteria.
- **Evaluation Dossier Viewer (`/report/:id`)**:
  - Scored criterion cards with composite technical scores and hiring recommendations.
  - Verbatim cited transcript evidence quotes proving competency.
  - Automatic polling for pending asynchronous evaluations.

---

## Tech Stack & Architecture

- **Framework**: Next.js 16 (App Router) with Turbopack
- **Language**: TypeScript 5.0+
- **Styling**: Tailwind CSS & CSS Variables
- **Icons**: Lucide React
- **Audio Capture**: Web Audio API (`AudioContext`, linear box downsampler, pre-roll buffer)

---

## Local Development

### 1. Prerequisites
- Node.js 20+ & npm

### 2. Configure Environment
Create `.env.local` in `frontend/`:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8003
NEXT_PUBLIC_WS_URL=ws://localhost:8004
```

### 3. Run Development Server
```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### 4. Production Build
```bash
npm run build
```

Turbopack compiles and statically prerenders all application routes.
