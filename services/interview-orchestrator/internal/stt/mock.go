package stt

import "context"

// DemoAdapter keeps the WebSocket audio contract active without requiring a
// paid speech provider. The browser still records and sends microphone audio.
type DemoAdapter struct{}

func NewDemoAdapter() Adapter { return DemoAdapter{} }

func (DemoAdapter) Transcribe(_ context.Context, _ []byte, _ string) (string, error) {
	return "I am ready to discuss my experience and approach.", nil
}
