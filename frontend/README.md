# VoxAssess — Web Application Frontend

The frontend interface for VoxAssess, built with **Next.js 14 (App Router)**, **TypeScript**, and **Tailwind CSS**. It provides a real-time conversational voice interview portal for candidates and an assessment management dashboard for hiring managers.

---

## Key Features & Capabilities

### 1. Real-Time Candidate Assessment Stage (`/interview/[token]`)
- **Web Audio Ingestion**: Captures linear 16kHz 16-bit PCM mono audio using standard Web Audio API and `AudioWorkletNode` buffers.
- **Full-Duplex WebSocket Streaming**: Streams binary PCM frames to `session-runtime` (`interview-orchestrator`) while receiving binary MP3 chunks and JSON control signals.
- **Audio Playback Queue & Visualizer**: Renders dynamic live waveforms and manages smooth audio playback with zero dead-air gaps between pipelined sentences.
- **Instant Barge-In Interruption**: Detects local speech during assessor playback, immediately halting the HTML5 AudioContext, clearing pending audio buffers, and dispatching interruption signals to the backend.

### 2. Hiring Manager Dashboard
- **Track & Blueprint Builder**: Author job tracks, custom technical interview prompts, time limits, and rubric scoring criteria.
- **Candidate Invite Management**: Generate single-use candidate invite links protected by unique tokens with configurable expiry.
- **Live Session Monitoring**: View active candidate connection statuses and turn progression.

### 3. Asynchronous Evaluation Report Viewer
- **Criterion Score Cards**: Visual breakdown of candidate competency ratings across rubric dimensions.
- **Verbatim Evidence Attribution**: Inspect exact conversation quotes cited by the AI evaluation worker to substantiate candidate scores.
- **Status Badges**: Real-time status indicators (`pending`, `evaluating`, `completed`, `evaluation_failed`).

---

## Tech Stack & Architecture

- **Framework**: Next.js 14.2+ (App Router)
- **Language**: TypeScript 5.0+
- **Styling**: Tailwind CSS & CSS Variables
- **Icons**: Lucide React
- **Audio Capture**: Web Audio API (`AudioContext`, `ScriptProcessorNode` / `AudioWorklet`)
- **Reverse Proxy**: Nginx (serving production assets and proxying `/api/` and `/ws/` requests)

---

## Local Development

### 1. Prerequisites
- Node.js 18+ & npm (or pnpm / yarn)

### 2. Configure Environment
Create `.env.local` in the `frontend` directory:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8001
NEXT_PUBLIC_WS_URL=ws://localhost:8002
```

### 3. Run Development Server
```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### 4. Production Build & Linting
```bash
npm run lint
npm run build
```
