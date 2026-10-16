package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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

type deepgramStreamMessage struct {
	Channel struct {
		Alternatives []struct {
			Transcript string `json:"transcript"`
		} `json:"alternatives"`
	} `json:"channel"`
	IsFinal     bool   `json:"is_final"`
	SpeechFinal bool   `json:"speech_final"`
	Type        string `json:"type"`
}

type deepgramStreamingSession struct {
	conn        *websocket.Conn
	resultsChan chan TranscriptResult
	sendMu      sync.Mutex
	closeOnce   sync.Once
	done        chan struct{}
}

func (a *deepgramAdapter) StreamingSession(ctx context.Context) (StreamingSession, error) {
	url := "wss://api.deepgram.com/v1/listen?model=nova-2&smart_format=true&punctuate=true&interim_results=true&endpointing=350&utterance_end_ms=700&vad_events=true&encoding=linear16&sample_rate=16000&channels=1"
	headers := http.Header{}
	headers.Set("Authorization", "Token "+a.apiKey)

	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		if resp != nil && resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("deepgram websocket dial failed (%d): %s: %w", resp.StatusCode, string(body), err)
		}
		return nil, fmt.Errorf("deepgram websocket dial failed: %w", err)
	}

	session := &deepgramStreamingSession{
		conn:        conn,
		resultsChan: make(chan TranscriptResult, 50),
		done:        make(chan struct{}),
	}

	// KeepAlive ticker (sends {"type":"KeepAlive"} every 8 seconds) to prevent 15s idle timeout
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-session.done:
				return
			case <-ticker.C:
				session.sendMu.Lock()
				_ = session.conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"KeepAlive"}`))
				session.sendMu.Unlock()
			}
		}
	}()

	go session.readLoop()

	return session, nil
}

func (s *deepgramStreamingSession) readLoop() {
	defer close(s.resultsChan)
	for {
		select {
		case <-s.done:
			return
		default:
			_, data, err := s.conn.ReadMessage()
			if err != nil {
				return
			}

			var msg deepgramStreamMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			if len(msg.Channel.Alternatives) > 0 {
				transcript := msg.Channel.Alternatives[0].Transcript
				if transcript != "" || msg.SpeechFinal || msg.IsFinal {
					select {
					case s.resultsChan <- TranscriptResult{
						Text:        transcript,
						IsFinal:     msg.IsFinal,
						SpeechFinal: msg.SpeechFinal,
					}:
					case <-s.done:
						return
					}
				}
			}
		}
	}
}

func (s *deepgramStreamingSession) SendAudio(ctx context.Context, audio []byte) error {
	select {
	case <-s.done:
		return fmt.Errorf("streaming session closed")
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.conn.WriteMessage(websocket.BinaryMessage, audio)
}

func (s *deepgramStreamingSession) Results() <-chan TranscriptResult {
	return s.resultsChan
}

func (s *deepgramStreamingSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.done)
		s.sendMu.Lock()
		_ = s.conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
		s.sendMu.Unlock()
		err = s.conn.Close()
	})
	return err
}
