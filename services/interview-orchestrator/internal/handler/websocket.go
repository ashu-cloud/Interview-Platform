package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/integration"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/kafka"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/session"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/stt"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/tts"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow any origin for now
	},
}

type WSMessage struct {
	Type        string `json:"type"`
	Data        string `json:"data,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
	Text        string `json:"text,omitempty"`
	Turn        int    `json:"turn,omitempty"`
	InterviewID string `json:"interview_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
}

type WebSocketHandler struct {
	jcsClient integration.JCSClient
	sessMgr   session.Manager
	sttClient stt.Adapter
	llmPool   *llm.Pool
	ttsClient tts.Adapter
	kafkaProd kafka.Producer
}

func NewWebSocketHandler(jcsClient integration.JCSClient, sessMgr session.Manager, sttClient stt.Adapter, llmPool *llm.Pool, ttsClient tts.Adapter, kafkaProd kafka.Producer) *WebSocketHandler {
	return &WebSocketHandler{
		jcsClient: jcsClient,
		sessMgr:   sessMgr,
		sttClient: sttClient,
		llmPool:   llmPool,
		ttsClient: ttsClient,
		kafkaProd: kafkaProd,
	}
}

func (h *WebSocketHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	h.handleConnection(ctx, conn, token)
}

func (h *WebSocketHandler) handleConnection(ctx context.Context, conn *websocket.Conn, token string) {
	sendError := func(code, msg string) {
		_ = conn.WriteJSON(WSMessage{
			Type:    "error",
			Code:    code,
			Message: msg,
		})
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg))
	}

	// 1. Validate Invite
	slog.Info("validating invite", "token", token)
	inviteRes, err := h.jcsClient.ValidateInvite(ctx, token)
	if err != nil {
		slog.Error("invite validation failed", "error", err)
		sendError("INVALID_INVITE", "Invalid or expired invite")
		return
	}
	if !inviteRes.Valid || inviteRes.InterviewStatus == "completed" {
		sendError("INVALID_INVITE", "Invalid or expired invite")
		return
	}

	interviewID := inviteRes.InterviewID
	lockValue := uuid.New().String()

	// 2. Acquire lock
	err = h.sessMgr.AcquireLock(ctx, interviewID, lockValue)
	if err != nil {
		if errors.Is(err, session.ErrLockFailed) {
			sendError("ALREADY_IN_PROGRESS", "Interview already in progress")
		} else {
			slog.Error("acquire lock failed", "error", err)
			sendError("INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	defer func() {
		_ = h.sessMgr.ReleaseLock(context.Background(), interviewID, lockValue)
	}()

	// Keep lock alive
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = h.sessMgr.RefreshLock(ctx, interviewID, lockValue)
			}
		}
	}()

	// 3. Load or Create Session
	sess, err := h.sessMgr.GetSession(ctx, interviewID)
	if err != nil && !errors.Is(err, session.ErrSessionNotFound) {
		slog.Error("get session failed", "error", err)
		sendError("INTERNAL_ERROR", "Internal server error")
		return
	}

	if sess == nil {
		// New Session
		docket, err := h.jcsClient.GetDocket(ctx, inviteRes.DocketID)
		if err != nil {
			slog.Error("failed to get docket", "error", err)
			sendError("INTERNAL_ERROR", "Internal server error")
			return
		}

		sess = &model.Session{
			InterviewID:         interviewID,
			CandidateID:         inviteRes.Candidate.ID,
			OrgID:               inviteRes.OrgID,
			DocketID:            inviteRes.DocketID,
			Docket:              *docket,
			ConversationHistory: []model.Message{},
			TurnNumber:          0,
			Status:              "in_progress",
			StartedAt:           time.Now().UTC(),
			LastActive:          time.Now().UTC(),
		}

		if err := h.sessMgr.SaveSession(ctx, sess); err != nil {
			slog.Error("save session failed", "error", err)
			sendError("INTERNAL_ERROR", "Internal server error")
			return
		}

		// Update interview status in DB
		changed, err := h.jcsClient.UpdateInterviewStatus(ctx, interviewID, "in_progress")
		if err != nil {
			slog.Error("update status failed", "error", err)
		}
		_ = changed // ignore for in_progress
	}

	slog.Info("session acquired", "interview_id", interviewID)

	_ = conn.WriteJSON(WSMessage{
		Type:        "connected",
		InterviewID: interviewID,
		SessionID:   interviewID,
	})

	// 4. If new session, send greeting and first question automatically
	if sess.TurnNumber == 0 && len(sess.ConversationHistory) == 0 {
		sess.TurnNumber = 1
		initialPrompt := "You are conducting an interview. Welcome the candidate warmly, introduce yourself briefly as the AI interviewer, and ask the first interview question to get started. IMPORTANT: Do NOT use any markdown formatting - no asterisks, no bold, no bullet points, no hashtags. Speak in plain conversational sentences only."
		initialHistory := []model.Message{
			{Role: "user", Content: initialPrompt},
		}
		_ = h.kafkaProd.PublishTurn(ctx, interviewID, sess.OrgID, 1, "user", initialPrompt)
		_ = h.sendAITurn(ctx, conn, sess, 1, h.buildSystemPrompt(&sess.Docket), initialHistory)
	}

	// 5. Turn loop
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			slog.Error("read message failed", "error", err)
			break
		}

		var req WSMessage
		if err := json.Unmarshal(msgBytes, &req); err != nil {
			slog.Error("unmarshal message failed", "error", err)
			continue
		}

		if req.Type == "heartbeat" {
			_ = conn.WriteJSON(WSMessage{Type: "heartbeat_ack"})
			continue
		}

		if req.Type == "audio_end" {
			audioBytes, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				slog.Error("decode audio failed", "error", err)
				continue
			}

			// 1. STT
			text, err := h.sttClient.Transcribe(ctx, audioBytes, req.MimeType)
			if err != nil {
				slog.Error("stt failed", "error", err)
				_ = conn.WriteJSON(WSMessage{
					Type:    "error",
					Code:    "STT_ERROR",
					Message: "Speech recognition had an issue. Please try speaking again.",
				})
				continue
			}

			// 2. Client events & Kafka
			sess.TurnNumber++
			turn := sess.TurnNumber

			_ = h.kafkaProd.PublishTurn(ctx, interviewID, sess.OrgID, turn, "user", text)
			sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
				Role: "user", Content: text,
			})

			// 3. AI response
			_ = h.sendAITurn(ctx, conn, sess, turn, h.buildSystemPrompt(&sess.Docket), sess.ConversationHistory)
		}

		if req.Type == "end_interview" {
			_ = conn.WriteJSON(WSMessage{Type: "interview_complete"})
			
			// Mark completed in JCS
			changed, err := h.jcsClient.UpdateInterviewStatus(ctx, interviewID, "completed")
			if err != nil {
				slog.Error("failed to update status to completed", "error", err)
			}
			
			// Only publish completed event if we actually transitioned the status
			if changed {
				_ = h.kafkaProd.PublishCompleted(ctx, interviewID, sess.OrgID, sess.CandidateID, sess.DocketID, time.Now().UTC(), sess.TurnNumber)
				
				// Clean up session
				_ = h.sessMgr.DeleteSession(ctx, interviewID)
			}
			break
		}
	}

	slog.Info("websocket connection closed", "interview_id", interviewID)
}

func (h *WebSocketHandler) sendAITurn(ctx context.Context, conn *websocket.Conn, sess *model.Session, turn int, systemPrompt string, history []model.Message) error {
	streamChan := make(chan string, 100)
	var llmText string

	go func() {
		defer close(streamChan)
		if err := h.llmPool.GenerateResponse(ctx, systemPrompt, history, streamChan); err != nil {
			slog.Error("llm failed", "error", err)
		}
	}()

	for chunk := range streamChan {
		llmText += chunk
		_ = conn.WriteJSON(WSMessage{
			Type: "ai_text_chunk",
			Turn: turn,
			Text: chunk,
		})
	}
	_ = conn.WriteJSON(WSMessage{Type: "ai_text_end", Turn: turn})

	if llmText == "" {
		return errors.New("empty LLM response")
	}

	// TTS (if TTS fails or errors out, ai_audio_end will still be sent so client falls back to browser synthesis)
	audioChan := make(chan []byte, 100)
	go func() {
		defer close(audioChan)
		if err := h.ttsClient.SynthesizeStream(ctx, llmText, audioChan); err != nil {
			slog.Error("tts failed", "error", err)
		}
	}()

	for audioChunk := range audioChan {
		_ = conn.WriteJSON(WSMessage{
			Type: "ai_audio_chunk",
			Turn: turn,
			Data: base64.StdEncoding.EncodeToString(audioChunk),
		})
	}
	_ = conn.WriteJSON(WSMessage{Type: "ai_audio_end", Turn: turn})

	_ = h.kafkaProd.PublishTurn(ctx, sess.InterviewID, sess.OrgID, turn, "assistant", llmText)
	sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
		Role: "assistant", Content: llmText,
	})

	sess.LastActive = time.Now().UTC()
	return h.sessMgr.SaveSession(ctx, sess)
}

func (h *WebSocketHandler) buildSystemPrompt(docket *model.Docket) string {
	var sb strings.Builder
	sb.WriteString("You are conducting a professional job interview.\n\n")
	if docket.Instructions != "" {
		sb.WriteString("Instructions:\n")
		sb.WriteString(docket.Instructions)
		sb.WriteString("\n\n")
	}
	if len(docket.Questions) > 0 {
		sb.WriteString("Interview Questions to cover:\n")
		for i, q := range docket.Questions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q.QuestionText))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("CRITICAL: Never use markdown formatting in your responses. No asterisks, no bold, no bullet points, no hashtags, no backticks. Respond in plain conversational sentences only - your text will be spoken aloud by a text-to-speech system.")
	return sb.String()
}
