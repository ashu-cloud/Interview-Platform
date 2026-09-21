package tts

import "context"

// DemoAdapter emits no provider audio. The browser uses SpeechSynthesis as a
// free local fallback after it receives the AI text response.
type DemoAdapter struct{}

func NewDemoAdapter() Adapter { return DemoAdapter{} }

func (DemoAdapter) SynthesizeStream(_ context.Context, _ string, _ chan<- []byte) error { return nil }
