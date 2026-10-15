package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Adapter interface {
	SynthesizeSentence(ctx context.Context, text string) ([]byte, error)
	SynthesizeStream(ctx context.Context, text string, audioChan chan<- []byte) error
}

type elevenLabsAdapter struct {
	apiKey  string
	voiceID string
	client  *http.Client
}

func NewElevenLabsAdapter(apiKey, voiceID string) Adapter {
	return &elevenLabsAdapter{
		apiKey:  apiKey,
		voiceID: voiceID,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (a *elevenLabsAdapter) SynthesizeSentence(ctx context.Context, text string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", a.voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_turbo_v2",
		"voice_settings": map[string]interface{}{
			"stability":        0.60,
			"similarity_boost": 0.80,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("xi-api-key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading audio body: %w", err)
	}

	return audioBytes, nil
}

func (a *elevenLabsAdapter) SynthesizeStream(ctx context.Context, text string, audioChan chan<- []byte) error {
	audioBytes, err := a.SynthesizeSentence(ctx, text)
	if err != nil {
		return err
	}
	if len(audioBytes) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case audioChan <- audioBytes:
		}
	}
	return nil
}
