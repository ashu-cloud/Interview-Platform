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
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/adaptive"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/integration"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/publisher"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/rag"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/session"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/storage"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/stt"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/tts"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow any origin for local dev / demo
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
	Sequence    int    `json:"sequence,omitempty"`
	IsFinal     bool   `json:"is_final,omitempty"`
	Speaking    bool   `json:"speaking,omitempty"`
}

type turnController struct {
	mu         sync.Mutex
	cancelFn   context.CancelFunc
	aiSpeaking bool
}

func (tc *turnController) startTurn() (context.Context, context.CancelFunc) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	tc.cancelFn = cancel
	tc.aiSpeaking = true
	return ctx, cancel
}

func (tc *turnController) interrupt() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.cancelFn != nil {
		tc.cancelFn()
	}
	tc.aiSpeaking = false
}

func (tc *turnController) done() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.aiSpeaking = false
}

func (tc *turnController) isSpeaking() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.aiSpeaking
}

type safeSender struct {
	writeChan chan WSMessage
	closed    bool
	mu        sync.Mutex
}

func newSafeSender(size int) *safeSender {
	return &safeSender{
		writeChan: make(chan WSMessage, size),
	}
}

func (s *safeSender) send(msg WSMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.writeChan <- msg:
	default:
		slog.Warn("writeChan full, message dropped", "type", msg.Type)
	}
}

func (s *safeSender) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.writeChan)
	}
}

type WebSocketHandler struct {
	jcsClient      integration.JCSClient
	sessMgr        session.Manager
	sttClient      stt.Adapter
	llmPool        *llm.Pool
	ttsClient      tts.Adapter
	eventPublisher publisher.Publisher
	audioOffloader *storage.AudioOffloader
	adaptiveEngine *adaptive.Engine
	ragPrefetcher  *rag.Prefetcher
	dbPool         *pgxpool.Pool
}

func NewWebSocketHandler(jcsClient integration.JCSClient, sessMgr session.Manager, sttClient stt.Adapter, llmPool *llm.Pool, ttsClient tts.Adapter, eventPublisher publisher.Publisher) *WebSocketHandler {
	return &WebSocketHandler{
		jcsClient:      jcsClient,
		sessMgr:        sessMgr,
		sttClient:      sttClient,
		llmPool:        llmPool,
		ttsClient:      ttsClient,
		eventPublisher: eventPublisher,
	}
}

func (h *WebSocketHandler) SetAudioOffloader(offloader *storage.AudioOffloader) {
	h.audioOffloader = offloader
}

func (h *WebSocketHandler) SetDBPool(pool *pgxpool.Pool) {
	h.dbPool = pool
}

// SetAdaptive wires the adaptive coverage engine and RAG prefetcher.
// Both may be nil, in which case turns use the static blueprint prompt.
func (h *WebSocketHandler) SetAdaptive(engine *adaptive.Engine, prefetcher *rag.Prefetcher) {
	h.adaptiveEngine = engine
	h.ragPrefetcher = prefetcher
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
	sender := newSafeSender(200)
	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		for msg := range sender.writeChan {
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()
	defer func() {
		sender.close()
		select {
		case <-writeDone:
		case <-time.After(500 * time.Millisecond):
		}
		_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}()

	send := sender.send

	var fatalErrMu sync.Mutex
	var hasFatalErr bool
	setFatalErr := func() {
		fatalErrMu.Lock()
		defer fatalErrMu.Unlock()
		hasFatalErr = true
	}
	isFatalErr := func() bool {
		fatalErrMu.Lock()
		defer fatalErrMu.Unlock()
		return hasFatalErr
	}

	sendError := func(code, msg string) {
		setFatalErr()
		send(WSMessage{
			Type:    "error",
			Code:    code,
			Message: msg,
		})
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

		// Adaptive engine: pre-computed criterion embeddings. Any failure or an
		// empty result disables adaptivity for this session without failing it.
		if h.adaptiveEngine != nil {
			embCtx, embCancel := context.WithTimeout(ctx, 5*time.Second)
			embeddings, embErr := h.jcsClient.GetCriteriaEmbeddings(embCtx, sess.DocketID)
			embCancel()
			if embErr != nil || len(embeddings) == 0 {
				slog.Warn("criteria embeddings unavailable, adaptive engine disabled",
					"blueprint_id", sess.DocketID, "error", embErr)
			} else {
				sess.CriteriaEmbeddings = embeddings
				sess.CoverageMatrix = make(map[string]float64, len(embeddings))
				for id := range embeddings {
					sess.CoverageMatrix[id] = 0.0
				}
			}
		}

		if err := h.sessMgr.SaveSession(ctx, sess); err != nil {
			slog.Error("save session failed", "error", err)
			sendError("INTERNAL_ERROR", "Internal server error")
			return
		}

		_, err = h.jcsClient.UpdateInterviewStatus(ctx, interviewID, "in_progress")
		if err != nil {
			slog.Error("update status failed", "error", err)
		}
	}

	slog.Info("session acquired", "interview_id", interviewID)

	send(WSMessage{
		Type:        "connected",
		InterviewID: interviewID,
		SessionID:   interviewID,
	})

	tc := &turnController{}

	// Turn Loop Channels & Concurrency Controls
	audioFrameChan := make(chan []byte, 300)
	speechEndedSignal := make(chan struct{}, 10)
	finalTranscriptChan := make(chan string, 10)
	connDone := make(chan struct{})

	var turnWg sync.WaitGroup
	var curTurn atomic.Int64
	curTurn.Store(int64(sess.TurnNumber))
	docketID := sess.DocketID

	var audioBufMu sync.Mutex
	var turnAudioBuffer []byte
	var currentMimeType = "audio/webm;codecs=opus"

	flushAudioBuffer := func(turnNum int) {
		audioBufMu.Lock()
		defer audioBufMu.Unlock()
		if len(turnAudioBuffer) > 0 && h.audioOffloader != nil {
			bufCopy := make([]byte, len(turnAudioBuffer))
			copy(bufCopy, turnAudioBuffer)
			h.audioOffloader.Enqueue(sess.InterviewID, turnNum, bufCopy, currentMimeType)
		}
		turnAudioBuffer = nil
	}

	var (
		lastInterimMu     sync.Mutex
		lastInterimText   string
		lastInterimTime   time.Time
		lastCommittedTurn int64 = -1
		turnCommitMu      sync.Mutex
	)

	commitTurn := func(text string, targetTurn int64) bool {
		turnCommitMu.Lock()
		defer turnCommitMu.Unlock()
		if lastCommittedTurn >= targetTurn {
			return false
		}
		clean := strings.TrimSpace(strings.Trim(text, ".,!?-~ "))
		words := strings.Fields(clean)
		if len(clean) <= 1 || len(words) == 0 || isFillerOnly(clean) {
			return false
		}
		lastCommittedTurn = targetTurn
		send(WSMessage{
			Type:    "final_transcript",
			Text:    clean,
			IsFinal: true,
			Turn:    int(targetTurn),
		})
		if tc.isSpeaking() {
			tc.interrupt()
			send(WSMessage{
				Type: "ai_interrupted",
				Turn: int(targetTurn),
			})
		}
		flushAudioBuffer(int(targetTurn))
		turnWg.Add(1)
		finalTranscriptChan <- clean
		return true
	}

	// Goroutine B — STT Session Manager with 900ms Speech Stability Auto-Commit Fallback
	go func() {
		var currentSession stt.StreamingSession
		var currentCancel context.CancelFunc

		closeCurrent := func() {
			if currentSession != nil {
				_ = currentSession.Close()
				currentSession = nil
			}
			if currentCancel != nil {
				currentCancel()
				currentCancel = nil
			}
		}
		defer closeCurrent()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-connDone:
				return
			case <-ticker.C:
				lastInterimMu.Lock()
				hasCandidate := lastInterimText != "" && time.Since(lastInterimTime) > 900*time.Millisecond
				candidateText := lastInterimText
				if hasCandidate {
					lastInterimText = ""
				}
				lastInterimMu.Unlock()

				if hasCandidate {
					target := curTurn.Load() + 1
					if commitTurn(candidateText, target) {
						slog.Info("stability timer committed speech turn", "interview_id", interviewID, "turn", target, "text", candidateText)
					}
				}

			case frame, ok := <-audioFrameChan:
				if !ok {
					return
				}
				if currentSession == nil {
					sttCtx, cancel := context.WithCancel(ctx)
					currentCancel = cancel
					s, err := h.sttClient.StreamingSession(sttCtx)
					if err != nil {
						slog.Error("stt streaming session start failed", "error", err)
						continue
					}
					currentSession = s

					// Listen for results
					go func(activeSession stt.StreamingSession) {
						for res := range activeSession.Results() {
							clean := strings.TrimSpace(strings.Trim(res.Text, ".,!?-~ "))
							if res.SpeechFinal {
								commitTurn(res.Text, curTurn.Load()+1)
								lastInterimMu.Lock()
								lastInterimText = ""
								lastInterimMu.Unlock()
							} else {
								send(WSMessage{
									Type:    "interim_transcript",
									Text:    res.Text,
									IsFinal: res.IsFinal,
								})
								if clean != "" {
									lastInterimMu.Lock()
									lastInterimText = clean
									lastInterimTime = time.Now()
									lastInterimMu.Unlock()

									if tc.isSpeaking() && !isFillerOnly(clean) {
										tc.interrupt()
										send(WSMessage{
											Type: "ai_interrupted",
											Turn: int(curTurn.Load()),
										})
									}
									h.ragPrefetcher.TriggerPrefetch(ctx, interviewID, docketID,
										int(curTurn.Load())+1, res.Text)
								}
							}
						}
					}(s)
				}

				if err := currentSession.SendAudio(ctx, frame); err != nil {
					slog.Warn("send audio to stt failed, resetting session for auto-reconnect", "error", err)
					closeCurrent()
				}

			case <-speechEndedSignal:
				closeCurrent()
			}
		}
	}()

	// 4. Initial Greeting on new or unstarted session (executed asynchronously so read loop is not blocked)
	if len(sess.ConversationHistory) == 0 {
		sess.TurnNumber = 1
		curTurn.Store(1)
		initialPrompt := "You are conducting an interview. Welcome the candidate warmly, introduce yourself briefly as the AI interviewer, and ask the first interview question to get started. IMPORTANT: Do NOT use any markdown formatting - no asterisks, no bold, no bullet points, no hashtags. Speak in plain conversational sentences only. Keep your response strictly under 3 sentences."
		initialHistory := []model.Message{
			{Role: "participant", Content: initialPrompt},
		}

		turnWg.Add(1)
		go func() {
			defer turnWg.Done()
			turnCtx, turnCancel := tc.startTurn()
			defer turnCancel()
			defer tc.done()
			if err := h.sendAITurn(turnCtx, send, sess, 1, buildSystemPrompt(&sess.Docket, initialHistory, "", "", ""), initialHistory); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("initial assessor turn failed", "interview_id", interviewID, "error", err)
			}
		}()
	}

	// Goroutine C — AI Response Generator
	var aiGeneratorStopped = make(chan struct{})
	go func() {
		defer close(aiGeneratorStopped)
		for {
			select {
			case <-connDone:
				return
			case text, ok := <-finalTranscriptChan:
				if !ok {
					return
				}
				if text == "" {
					turnWg.Done()
					continue
				}

				sess.TurnNumber++
				turn := sess.TurnNumber
				curTurn.Store(int64(turn))

				// RAG lookup runs in parallel with the coverage update; it is served
				// from the interim-transcript prefetch cache when available.
				ragCh := make(chan string, 1)
				go func(finalText string, n int) {
					ragCtx, _ := h.ragPrefetcher.GetContext(ctx, interviewID, docketID, n, finalText)
					ragCh <- ragCtx
				}(text, turn)

				// 1. Update the coverage matrix from the participant's answer.
				adaptiveOn := h.adaptiveEngine != nil && len(sess.CriteriaEmbeddings) > 0
				if adaptiveOn {
					if sess.CoverageMatrix == nil {
						sess.CoverageMatrix = make(map[string]float64, len(sess.CriteriaEmbeddings))
						for id := range sess.CriteriaEmbeddings {
							sess.CoverageMatrix[id] = 0.0
						}
					}
					covCtx, covCancel := context.WithTimeout(ctx, 3*time.Second)
					updated, covErr := h.adaptiveEngine.UpdateCoverage(covCtx, text,
						sess.CriteriaEmbeddings, sess.CoverageMatrix)
					covCancel()
					if covErr != nil {
						slog.Warn("coverage update failed", "interview_id", interviewID, "error", covErr)
					} else {
						sess.CoverageMatrix = updated
					}
				}

				sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
					Role: "participant", Content: text, Timestamp: time.Now().UTC(),
				})
				sess.LastActive = time.Now().UTC()
				if err := h.sessMgr.SaveSession(ctx, sess); err != nil {
					slog.Error("save participant turn failed", "interview_id", interviewID, "error", err)
					sendError("INTERNAL_ERROR", "Internal server error")
					turnWg.Done()
					return
				}
				h.publishTurn(ctx, interviewID, sess.OrgID, turn, publisher.RoleParticipant, text)

				// 2. Pick the criterion with the largest coverage deficit.
				var focusName, focusDesc string
				if adaptiveOn {
					focusName, focusDesc = h.adaptiveEngine.NextDirective(sess.CoverageMatrix, sess.Docket.Rubric)
				}

				// 3. Collect RAG context (never blocks the turn past the query budget).
				var ragContext string
				select {
				case ragContext = <-ragCh:
				case <-ctx.Done():
				}

				// 4. Build the system prompt with directive + context.
				systemPrompt := buildSystemPrompt(&sess.Docket, sess.ConversationHistory,
					focusName, focusDesc, ragContext)
				slog.Info("assessor prompt built",
					"interview_id", interviewID, "turn", turn,
					"current_focus", focusName, "rag_context_chars", len(ragContext),
					"coverage", sess.CoverageMatrix)

				turnCtx, turnCancel := tc.startTurn()
				err := h.sendAITurn(turnCtx, send, sess, turn, systemPrompt, sess.ConversationHistory)
				turnCancel()
				tc.done()

				if err != nil && !errors.Is(err, context.Canceled) {
					slog.Error("assessor turn failed", "interview_id", interviewID, "error", err)
					sendError("INTERNAL_ERROR", "Internal server error")
					turnWg.Done()
					return
				}
				turnWg.Done()
			}
		}
	}()

	// Goroutine A — Main Read Loop
	for {
		if isFatalErr() {
			break
		}
		_, msgBytes, err := conn.ReadMessage()
		if err != nil || isFatalErr() {
			slog.Info("websocket connection closed", "interview_id", interviewID)
			break
		}

		var req WSMessage
		if err := json.Unmarshal(msgBytes, &req); err != nil {
			slog.Error("unmarshal message failed", "error", err)
			continue
		}

		switch req.Type {
		case "audio_frame":
			if req.MimeType != "" {
				audioBufMu.Lock()
				currentMimeType = req.MimeType
				audioBufMu.Unlock()
			}
			data, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				slog.Error("decode audio frame failed", "error", err)
				continue
			}
			audioBufMu.Lock()
			turnAudioBuffer = append(turnAudioBuffer, data...)
			audioBufMu.Unlock()
			select {
			case audioFrameChan <- data:
			default:
				slog.Warn("audioFrameChan full, dropping frame")
			}

		case "user_interrupted", "speech_started":
			if tc.isSpeaking() {
				tc.interrupt()
				send(WSMessage{
					Type: "ai_interrupted",
					Turn: sess.TurnNumber,
				})
			}

		case "reset_session":
			slog.Info("reset_session requested", "interview_id", interviewID)
			tc.interrupt()
			turnCommitMu.Lock()
			lastCommittedTurn = 0
			turnCommitMu.Unlock()
			lastInterimMu.Lock()
			lastInterimText = ""
			lastInterimMu.Unlock()
			_ = h.sessMgr.DeleteSession(ctx, interviewID)
			if h.dbPool != nil {
				_, dbErr := h.dbPool.Exec(ctx,
					"UPDATE assessment_sessions SET status = 'in_progress', completed_at = NULL WHERE id = $1",
					interviewID)
				if dbErr != nil {
					slog.Warn("reset_session: db assessment_sessions reset failed", "error", dbErr)
				}
				_, turnsErr := h.dbPool.Exec(ctx,
					"DELETE FROM session_turns WHERE session_id = $1",
					interviewID)
				if turnsErr != nil {
					slog.Warn("reset_session: db session_turns delete failed", "error", turnsErr)
				}
			}
			sess.ConversationHistory = []model.Message{}
			sess.TurnNumber = 1
			curTurn.Store(1)
			_ = h.sessMgr.SaveSession(ctx, sess)

			send(WSMessage{
				Type: "session_reset",
				Turn: 1,
			})

			// Trigger fresh initial greeting turn
			initialPrompt := "You are conducting an interview. Welcome the candidate warmly, introduce yourself briefly as the AI interviewer, and ask the first interview question to get started. IMPORTANT: Do NOT use any markdown formatting - no asterisks, no bold, no bullet points, no hashtags. Speak in plain conversational sentences only. Keep your response strictly under 3 sentences."
			initialHistory := []model.Message{
				{Role: "participant", Content: initialPrompt},
			}
			turnWg.Add(1)
			go func() {
				defer turnWg.Done()
				turnCtx, turnCancel := tc.startTurn()
				defer turnCancel()
				defer tc.done()
				if err := h.sendAITurn(turnCtx, send, sess, 1, buildSystemPrompt(&sess.Docket, initialHistory, "", "", ""), initialHistory); err != nil && !errors.Is(err, context.Canceled) {
					slog.Error("reset greeting turn failed", "interview_id", interviewID, "error", err)
				}
			}()

		case "speech_ended":
			flushAudioBuffer(sess.TurnNumber + 1)
			select {
			case speechEndedSignal <- struct{}{}:
			default:
			}

		case "audio_end":
			// Legacy full-blob path
			data, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				slog.Error("decode audio failed", "error", err)
				continue
			}
			if h.audioOffloader != nil {
				h.audioOffloader.Enqueue(sess.InterviewID, sess.TurnNumber+1, data, req.MimeType)
			}

			text, err := h.sttClient.Transcribe(ctx, data, req.MimeType)
			if err != nil {
				slog.Error("stt failed", "error", err)
				send(WSMessage{
					Type:    "error",
					Code:    "STT_ERROR",
					Message: "Speech recognition error",
				})
				continue
			}
			turnWg.Add(1)
			finalTranscriptChan <- text

		case "heartbeat":
			send(WSMessage{Type: "heartbeat_ack"})

		case "end_assessment", "end_interview":
			// Wait for any in-flight speech turn to finish processing
			turnWg.Wait()
			if isFatalErr() {
				close(connDone)
				return
			}

			sess.Status = "completed"
			sess.LastActive = time.Now().UTC()
			if err := h.sessMgr.SaveSession(ctx, sess); err != nil {
				slog.Error("save completed session failed", "interview_id", interviewID, "error", err)
				sendError("INTERNAL_ERROR", "Internal server error")
				close(connDone)
				return
			}
			send(WSMessage{Type: "assessment_complete"})
			send(WSMessage{Type: "interview_complete"}) // legacy compatibility

			changed, err := h.jcsClient.UpdateInterviewStatus(ctx, interviewID, "completed")
			if err != nil {
				slog.Error("failed to update status to completed", "error", err)
			}
			if changed {
				h.publishCompleted(ctx, interviewID, sess.OrgID, sess.CandidateID, sess.DocketID, sess.TurnNumber)
				_ = h.sessMgr.DeleteSession(ctx, interviewID)
			}
			close(connDone)
			return
		}
	}

	close(connDone)
}

func findSentenceEnd(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' {
			return i
		}
		if c == '.' || c == '!' || c == '?' {
			// Don't break on decimals (e.g. 3.14) or version numbers (e.g. 1.0)
			if c == '.' && i > 0 && i+1 < len(s) {
				if s[i-1] >= '0' && s[i-1] <= '9' && s[i+1] >= '0' && s[i+1] <= '9' {
					continue
				}
			}
			// Sentence break requires trailing whitespace or newline
			if i+1 < len(s) && (s[i+1] == ' ' || s[i+1] == '\t' || s[i+1] == '\r' || s[i+1] == '\n') {
				return i
			}
		}
	}
	// Fallback for long phrases without terminal punctuation (>120 chars)
	if len(s) > 120 {
		if idx := strings.LastIndexAny(s, ",;"); idx >= 50 {
			return idx
		}
		if idx := strings.LastIndexByte(s, ' '); idx >= 50 {
			return idx
		}
	}
	return -1
}

func (h *WebSocketHandler) sendAITurn(ctx context.Context, send func(WSMessage),
	sess *model.Session, turnNum int, systemPrompt string,
	history []model.Message) error {

	send(WSMessage{
		Type:     "ai_speaking",
		Speaking: true,
		Turn:     turnNum,
	})

	streamChan := make(chan string, 100)
	go func() {
		defer close(streamChan)
		_ = h.llmPool.GenerateResponse(ctx, systemPrompt, history, streamChan)
	}()

	var fullText strings.Builder
	var sentenceBuf strings.Builder

	sentenceChan := make(chan string, 32)
	var ttsWg sync.WaitGroup
	ttsWg.Add(1)
	go func() {
		defer ttsWg.Done()
		seq := 0
		for {
			select {
			case <-ctx.Done():
				return
			case sentence, ok := <-sentenceChan:
				if !ok {
					return
				}
				trimmed := strings.TrimSpace(sentence)
				if trimmed == "" {
					continue
				}
				audioBytes, err := h.ttsClient.SynthesizeSentence(ctx, trimmed)
				if err != nil {
					if ctx.Err() == nil {
						slog.Error("tts sentence synthesis failed", "turn", turnNum, "error", err, "sentence", trimmed)
					}
					continue
				}
				if len(audioBytes) > 0 && ctx.Err() == nil {
					seq++
					send(WSMessage{
						Type:     "ai_audio_chunk",
						Turn:     turnNum,
						Sequence: seq,
						Data:     base64.StdEncoding.EncodeToString(audioBytes),
					})
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			send(WSMessage{
				Type: "ai_interrupted",
				Turn: turnNum,
			})
			send(WSMessage{
				Type:     "ai_speaking",
				Speaking: false,
				Turn:     turnNum,
			})
			if fullText.Len() > 0 {
				sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
					Role:      "assessor",
					Content:   fullText.String() + " [interrupted]",
					Timestamp: time.Now().UTC(),
				})
				sess.LastActive = time.Now().UTC()
				_ = h.sessMgr.SaveSession(context.Background(), sess)
			}
			return ctx.Err()

		case chunk, ok := <-streamChan:
			if !ok {
				// Flush remaining sentence buffer
				if rem := strings.TrimSpace(sentenceBuf.String()); rem != "" {
					select {
					case <-ctx.Done():
					case sentenceChan <- rem:
					}
				}
				close(sentenceChan)
				ttsWg.Wait()
				send(WSMessage{Type: "ai_text_end", Turn: turnNum})
				send(WSMessage{Type: "ai_audio_end", Turn: turnNum})
				goto done
			}
			if chunk == "" {
				continue
			}
			fullText.WriteString(chunk)
			sentenceBuf.WriteString(chunk)
			send(WSMessage{Type: "ai_text_chunk", Text: chunk, Turn: turnNum})

			for {
				bufStr := sentenceBuf.String()
				endIdx := findSentenceEnd(bufStr)
				if endIdx < 0 {
					break
				}
				sentence := strings.TrimSpace(bufStr[:endIdx+1])
				remaining := strings.TrimLeft(bufStr[endIdx+1:], " \t\r\n")
				sentenceBuf.Reset()
				sentenceBuf.WriteString(remaining)

				if sentence != "" {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case sentenceChan <- sentence:
					}
				}
			}
		}
	}

done:
	if fullText.Len() == 0 {
		return errors.New("empty LLM response")
	}

	sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
		Role: "assessor", Content: fullText.String(), Timestamp: time.Now().UTC(),
	})
	sess.LastActive = time.Now().UTC()
	if err := h.sessMgr.SaveSession(ctx, sess); err != nil {
		return err
	}
	h.publishTurn(ctx, sess.InterviewID, sess.OrgID, turnNum, publisher.RoleAssessor, fullText.String())
	return nil
}

func (h *WebSocketHandler) pipeToTTS(ctx context.Context, send func(WSMessage), sentence string, turnNum int) {
	trimmed := strings.TrimSpace(sentence)
	if trimmed == "" {
		return
	}
	audioBytes, err := h.ttsClient.SynthesizeSentence(ctx, trimmed)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("tts synthesis failed", "error", err)
		}
		return
	}
	if len(audioBytes) > 0 && ctx.Err() == nil {
		send(WSMessage{
			Type: "ai_audio_chunk",
			Turn: turnNum,
			Data: base64.StdEncoding.EncodeToString(audioBytes),
		})
	}
}

func (h *WebSocketHandler) publishTurn(ctx context.Context, sessionID, orgID string, turn int, role, content string) {
	if err := h.eventPublisher.PublishTurn(ctx, sessionID, orgID, turn, role, content); err != nil {
		slog.Warn("best-effort turn publication failed", "session_id", sessionID, "role", role, "error", err)
	}
}

func (h *WebSocketHandler) publishCompleted(ctx context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) {
	if err := h.eventPublisher.PublishCompleted(ctx, sessionID, orgID, participantID, blueprintID, turnCount); err != nil {
		slog.Warn("best-effort completion publication failed", "session_id", sessionID, "error", err)
	}
}

// buildSystemPrompt renders the static blueprint prompt and, when present,
// appends the RAG context and the adaptive CURRENT FOCUS directive.
func buildSystemPrompt(docket *model.Docket, _ []model.Message,
	focusName, focusDesc, ragContext string) string {
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
	if ragContext != "" {
		sb.WriteString(ragContext)
		sb.WriteString("\n")
	}
	if focusName != "" {
		sb.WriteString(fmt.Sprintf(
			"CURRENT FOCUS: The participant has not yet adequately demonstrated "+
				"[%s] (%s). Your next question MUST probe this competency specifically.\n\n",
			focusName, focusDesc))
	}
	sb.WriteString("CRITICAL VOICE & BREVITY CONSTRAINTS:\n" +
		"- Never use markdown formatting in your responses. No asterisks, no bold, no bullet points, no hashtags, no backticks. Respond in plain conversational spoken sentences only.\n" +
		"- BREVITY IS MANDATORY: Speak strictly 2 to 3 concise sentences (under 45 words total). Keep questions clear, focused, and conversational. Do NOT give long preambles, explanations, or multiple sub-questions.\n")
	return sb.String()
}

func isFillerOnly(s string) bool {
	clean := strings.ToLower(strings.TrimSpace(strings.Trim(s, ".,!?-~ ")))
	fillers := map[string]bool{
		"um": true, "uh": true, "ah": true, "er": true, "hmm": true,
		"like": true, "so": true, "yeah": true, "ok": true, "okay": true,
	}
	return fillers[clean]
}

