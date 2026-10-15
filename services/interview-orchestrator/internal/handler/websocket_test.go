package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/integration"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/stt"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/tts"
)

type testJCS struct{ integration.JCSClient }

func (testJCS) ValidateInvite(context.Context, string) (*integration.InviteValidateResponse, error) {
	return &integration.InviteValidateResponse{Valid: true, InterviewID: "session"}, nil
}

func (testJCS) UpdateInterviewStatus(context.Context, string, string) (bool, error) {
	return true, nil
}

type testSessionManager struct {
	trace    *[]string
	failSave int
	saves    int
}

func (*testSessionManager) AcquireLock(context.Context, string, string) error { return nil }
func (*testSessionManager) RefreshLock(context.Context, string, string) error { return nil }
func (*testSessionManager) ReleaseLock(context.Context, string, string) error { return nil }
func (*testSessionManager) GetSession(context.Context, string) (*model.Session, error) {
	return &model.Session{
		InterviewID: "session", OrgID: "org", CandidateID: "participant", DocketID: "blueprint",
		Status: "in_progress",
	}, nil
}
func (m *testSessionManager) SaveSession(_ context.Context, sess *model.Session) error {
	m.saves++
	label := "save:completed"
	if sess.Status != "completed" {
		label = "save:" + sess.ConversationHistory[len(sess.ConversationHistory)-1].Role
	}
	*m.trace = append(*m.trace, label)
	if m.saves == m.failSave {
		return errors.New("Redis unavailable")
	}
	return nil
}
func (m *testSessionManager) DeleteSession(context.Context, string) error {
	*m.trace = append(*m.trace, "delete")
	return nil
}

type testPublisher struct {
	trace *[]string
	err   error
}

func (p *testPublisher) PublishTurn(_ context.Context, sessionID, orgID string, turn int, role, content string) error {
	*p.trace = append(*p.trace, fmt.Sprintf("publish:%s:%d", role, turn))
	return p.err
}
func (p *testPublisher) PublishCompleted(_ context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) error {
	*p.trace = append(*p.trace, "publish:completed")
	return p.err
}
func (*testPublisher) Close() error { return nil }

func TestWebSocketPersistsBeforeBestEffortPublishing(t *testing.T) {
	fullTrace := []string{
		"save:assessor", "publish:assessor:1",
		"save:participant", "publish:participant:2",
		"save:assessor", "publish:assessor:2",
		"save:completed", "publish:completed", "delete",
	}
	for _, tc := range []struct {
		name     string
		failSave int
		pubErr   error
		want     []string
	}{
		{name: "successful session", want: fullTrace},
		{name: "publishing errors remain best effort", pubErr: errors.New("publisher unavailable"), want: fullTrace},
		{name: "initial assessor save failure", failSave: 1, want: fullTrace[:1]},
		{name: "participant save failure", failSave: 2, want: fullTrace[:3]},
		{name: "assessor save failure", failSave: 3, want: fullTrace[:5]},
		{name: "completion save failure", failSave: 4, want: fullTrace[:7]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var trace []string
			mgr := &testSessionManager{trace: &trace, failSave: tc.failSave}
			pub := &testPublisher{trace: &trace, err: tc.pubErr}
			demo := llm.NewDemoProvider()
			h := NewWebSocketHandler(testJCS{}, mgr, stt.NewDemoAdapter(), llm.NewPool(demo, demo), tts.NewDemoAdapter(), pub)
			done := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				h.HandleWS(w, r)
			}))
			defer srv.Close()

			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/interview/?token=test", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			// Legacy client messages are queued and processed after the initial greeting.
			for _, msg := range []WSMessage{{Type: "audio_end", Data: "YQ==", MimeType: "audio/webm"}, {Type: "end_interview"}} {
				if err := conn.WriteJSON(msg); err != nil {
					t.Fatal(err)
				}
			}
			sawComplete, sawError := false, false
			for {
				var msg WSMessage
				if err := conn.ReadJSON(&msg); err != nil {
					break
				}
				sawComplete = sawComplete || msg.Type == "interview_complete"
				sawError = sawError || msg.Type == "error"
			}
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("websocket handler did not terminate")
			}
			if !reflect.DeepEqual(trace, tc.want) {
				t.Fatalf("trace = %v, want %v", trace, tc.want)
			}
			if tc.failSave == 0 && !sawComplete {
				t.Fatal("missing legacy interview_complete message")
			}
			if tc.failSave != 0 && (!sawError || sawComplete) {
				t.Fatalf("save failure: sawError=%v sawComplete=%v", sawError, sawComplete)
			}
		})
	}
}

type slowStreamingLLM struct{}

func (slowStreamingLLM) GenerateResponse(ctx context.Context, _ string, _ []model.Message, streamChan chan<- string) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(60 * time.Millisecond):
			streamChan <- fmt.Sprintf("word %d. ", i)
		}
	}
	return nil
}

func TestStreamingAudioAndInterruption(t *testing.T) {
	var trace []string
	mgr := &testSessionManager{trace: &trace}
	pub := &testPublisher{trace: &trace}
	slow := slowStreamingLLM{}
	h := NewWebSocketHandler(testJCS{}, mgr, stt.NewDemoAdapter(), llm.NewPool(slow, slow), tts.NewDemoAdapter(), pub)

	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		h.HandleWS(w, r)
	}))
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/interview/?token=test", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	sawConnected := false
	sawGreeting := false
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read greeting: %v", err)
		}
		if msg.Type == "connected" {
			sawConnected = true
		}
		if msg.Type == "ai_text_end" {
			sawGreeting = true
			break
		}
	}
	if !sawConnected || !sawGreeting {
		t.Fatalf("expected connected and greeting, got sawConnected=%v, sawGreeting=%v", sawConnected, sawGreeting)
	}

	// 1. Participant streams speech
	_ = conn.WriteJSON(WSMessage{Type: "speech_started"})
	_ = conn.WriteJSON(WSMessage{
		Type:     "audio_frame",
		Data:     "YQ==",
		Sequence: 1,
		MimeType: "audio/webm;codecs=opus",
	})
	_ = conn.WriteJSON(WSMessage{Type: "speech_ended"})

	// 2. Expect final_transcript
	sawFinalTranscript := false
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read final transcript: %v", err)
		}
		if msg.Type == "final_transcript" {
			sawFinalTranscript = true
			break
		}
	}
	if !sawFinalTranscript {
		t.Fatal("missing final_transcript message")
	}

	// 3. Wait for AI response to start, then interrupt!
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read ai turn: %v", err)
		}
		if msg.Type == "ai_text_chunk" {
			_ = conn.WriteJSON(WSMessage{Type: "speech_started"})
			break
		}
	}

	// 4. Expect ai_interrupted
	sawInterrupted := false
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Type == "ai_interrupted" {
			sawInterrupted = true
			break
		}
	}
	if !sawInterrupted {
		t.Fatal("missing ai_interrupted message")
	}

	// 5. Conclude session
	_ = conn.WriteJSON(WSMessage{Type: "end_assessment"})
	sawComplete := false
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Type == "assessment_complete" {
			sawComplete = true
			break
		}
	}
	if !sawComplete {
		t.Fatal("missing assessment_complete message")
	}
}

