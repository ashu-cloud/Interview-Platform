package stt

import "context"

// Adapter defines the speech-to-text contract supporting both one-shot and streaming transcription.
type Adapter interface {
	// Transcribe is the one-shot path. Still used in demo mode or fallback.
	Transcribe(ctx context.Context, audio []byte, mimeType string) (string, error)
	// StreamingSession opens a real-time transcription session.
	StreamingSession(ctx context.Context) (StreamingSession, error)
}

// TranscriptResult represents a partial or finalized speech-to-text segment.
type TranscriptResult struct {
	Text        string
	IsFinal     bool // true when Deepgram commits words (interim chunk)
	SpeechFinal bool // true ONLY when endpointing silence fires (utterance boundary)
}

// StreamingSession handles bidirectional real-time audio streaming and transcription.
type StreamingSession interface {
	SendAudio(ctx context.Context, audio []byte) error
	Results() <-chan TranscriptResult
	Close() error
}
