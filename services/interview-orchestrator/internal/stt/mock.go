package stt

import (
	"context"
	"sync"
)

// DemoAdapter keeps the WebSocket audio contract active without requiring a
// paid speech provider. The browser still records and sends microphone audio.
type DemoAdapter struct{}

func NewDemoAdapter() Adapter { return DemoAdapter{} }

func (DemoAdapter) Transcribe(_ context.Context, _ []byte, _ string) (string, error) {
	return "I am ready to discuss my experience and approach.", nil
}

func (DemoAdapter) StreamingSession(_ context.Context) (StreamingSession, error) {
	return &demoStreamingSession{
		resultsChan: make(chan TranscriptResult, 1),
	}, nil
}

type demoStreamingSession struct {
	resultsChan chan TranscriptResult
	closeOnce   sync.Once
}

func (d *demoStreamingSession) SendAudio(_ context.Context, _ []byte) error {
	// Discard all incoming audio frames in demo mode
	return nil
}

func (d *demoStreamingSession) Results() <-chan TranscriptResult {
	return d.resultsChan
}

func (d *demoStreamingSession) Close() error {
	d.closeOnce.Do(func() {
		d.resultsChan <- TranscriptResult{
			Text:    "I am ready to discuss my experience.",
			IsFinal: true,
		}
		close(d.resultsChan)
	})
	return nil
}
