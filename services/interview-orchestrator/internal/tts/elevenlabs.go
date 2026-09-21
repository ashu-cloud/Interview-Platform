package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Adapter interface {
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
		client:  &http.Client{},
	}
}

func (a *elevenLabsAdapter) SynthesizeStream(ctx context.Context, text string, audioChan chan<- []byte) error {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream", a.voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_turbo_v2",
		"voice_settings": map[string]float64{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("xi-api-key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			audioChan <- chunk
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}
	}

	return nil
}
