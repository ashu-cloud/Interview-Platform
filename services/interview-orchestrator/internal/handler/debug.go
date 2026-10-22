package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/config"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/tts"
)

type HopCheck struct {
	Hop        string `json:"hop"`
	OK         bool   `json:"ok"`
	DurationMS int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type DebugVoiceResponse struct {
	OK        bool       `json:"ok"`
	Timestamp string     `json:"timestamp"`
	Checks    []HopCheck `json:"checks"`
}

type DebugHandler struct {
	cfg         *config.Config
	redisClient *redis.Client
	llmPool     *llm.Pool
	ttsAdapter  tts.Adapter
}

func NewDebugHandler(cfg *config.Config, redisClient *redis.Client, llmPool *llm.Pool, ttsAdapter tts.Adapter) *DebugHandler {
	return &DebugHandler{
		cfg:         cfg,
		redisClient: redisClient,
		llmPool:     llmPool,
		ttsAdapter:  ttsAdapter,
	}
}

func (h *DebugHandler) HandleDebugVoice(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VoiceDebug {
		http.NotFound(w, r)
		return
	}

	key := r.Header.Get("X-Internal-Key")
	if key == "" || key != h.cfg.InternalAPIKey {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	checks := make([]HopCheck, 0, 5)
	allOK := true

	// 1. Redis Check
	t0 := time.Now()
	err := h.redisClient.Ping(ctx).Err()
	d0 := time.Since(t0).Milliseconds()
	check0 := HopCheck{Hop: "redis", OK: err == nil, DurationMS: d0}
	if err != nil {
		check0.Error = err.Error()
		allOK = false
	}
	checks = append(checks, check0)

	// 2. Platform Service HTTP Health
	t1 := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, h.cfg.JobCandidateServiceURL+"/health", nil)
	resp, err := client.Do(req)
	d1 := time.Since(t1).Milliseconds()
	check1 := HopCheck{Hop: "platform_service", OK: false, DurationMS: d1}
	if err != nil {
		check1.Error = err.Error()
		allOK = false
	} else {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			check1.OK = true
		} else {
			check1.Error = fmt.Sprintf("status %d", resp.StatusCode)
			allOK = false
		}
	}
	checks = append(checks, check1)

	// 3. Deepgram STT Connection Probe
	t2 := time.Now()
	check2 := HopCheck{Hop: "deepgram_stt", OK: false}
	if h.cfg.DemoMode {
		check2.OK = true
		check2.DurationMS = 0
	} else if h.cfg.DeepgramAPIKey == "" {
		check2.Error = "DEEPGRAM_API_KEY is not set"
		allOK = false
	} else {
		dgURL := "wss://api.deepgram.com/v1/listen?model=nova-2&encoding=linear16&sample_rate=16000&channels=1"
		hdr := http.Header{}
		hdr.Set("Authorization", "Token "+h.cfg.DeepgramAPIKey)
		dialer := websocket.DefaultDialer
		dgConn, dgResp, dgErr := dialer.DialContext(ctx, dgURL, hdr)
		check2.DurationMS = time.Since(t2).Milliseconds()
		if dgErr != nil {
			if dgResp != nil {
				check2.Error = fmt.Sprintf("dial failed (HTTP %d): %v", dgResp.StatusCode, dgErr)
			} else {
				check2.Error = dgErr.Error()
			}
			allOK = false
		} else {
			_ = dgConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
			_ = dgConn.Close()
			check2.OK = true
		}
	}
	checks = append(checks, check2)

	// 4. Groq LLM Completion Probe
	t3 := time.Now()
	check3 := HopCheck{Hop: "groq_llm", OK: false}
	if h.cfg.DemoMode {
		check3.OK = true
		check3.DurationMS = 0
	} else if h.cfg.GroqAPIKey == "" {
		check3.Error = "GROQ_API_KEY is not set"
		allOK = false
	} else {
		llmCh := make(chan string, 10)
		llmErrCh := make(chan error, 1)
		go func() {
			err := h.llmPool.GenerateResponse(ctx, "Answer with one word: ping", nil, llmCh)
			llmErrCh <- err
		}()
		// Drain stream
		go func() {
			for range llmCh {
			}
		}()
		select {
		case err := <-llmErrCh:
			check3.DurationMS = time.Since(t3).Milliseconds()
			if err != nil {
				check3.Error = err.Error()
				allOK = false
			} else {
				check3.OK = true
			}
		case <-time.After(5 * time.Second):
			check3.DurationMS = time.Since(t3).Milliseconds()
			check3.Error = "groq timeout after 5s"
			allOK = false
		}
	}
	checks = append(checks, check3)

	// 5. ElevenLabs TTS Synthesis Probe
	t4 := time.Now()
	check4 := HopCheck{Hop: "elevenlabs_tts", OK: false}
	if h.cfg.DemoMode {
		check4.OK = true
		check4.DurationMS = 0
	} else if h.cfg.ElevenLabsAPIKey == "" {
		check4.Error = "ELEVENLABS_API_KEY is not set"
		allOK = false
	} else {
		audio, err := h.ttsAdapter.SynthesizeSentence(ctx, "Test.")
		check4.DurationMS = time.Since(t4).Milliseconds()
		if err != nil {
			check4.Error = err.Error()
			allOK = false
		} else if len(audio) == 0 {
			check4.Error = "empty audio returned"
			allOK = false
		} else {
			check4.OK = true
		}
	}
	checks = append(checks, check4)

	res := DebugVoiceResponse{
		OK:        allOK,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if !allOK {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(res)
}
