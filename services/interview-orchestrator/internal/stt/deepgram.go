package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Adapter interface {
	Transcribe(ctx context.Context, audioBytes []byte, mimeType string) (string, error)
}

type deepgramAdapter struct {
	apiKey string
	client *http.Client
}

func NewDeepgramAdapter(apiKey string) Adapter {
	return &deepgramAdapter{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type deepgramResponse struct {
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string `json:"transcript"`
			} `json:"alternatives"`
		} `json:"channels"`
	} `json:"results"`
}

func (a *deepgramAdapter) Transcribe(ctx context.Context, audioBytes []byte, mimeType string) (string, error) {
	url := "https://api.deepgram.com/v1/listen?model=nova-2&smart_format=true&punctuate=true&language=en-US"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(audioBytes))
	if err != nil {
		return "", fmt.Errorf("create deepgram request failed: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.apiKey)
	if mimeType == "" {
		mimeType = "audio/webm"
	}
	req.Header.Set("Content-Type", mimeType)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do deepgram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepgram error %d: %s", resp.StatusCode, string(b))
	}

	var res deepgramResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode deepgram response failed: %w", err)
	}

	if len(res.Results.Channels) > 0 && len(res.Results.Channels[0].Alternatives) > 0 {
		return res.Results.Channels[0].Alternatives[0].Transcript, nil
	}

	return "", nil
}
