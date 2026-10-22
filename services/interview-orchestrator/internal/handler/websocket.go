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
	HasHistory  bool   `json:"has_history,omitempty"`
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
	voiceDebug     bool
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

func (h *WebSocketHandler) SetVoiceDebug(debug bool) {
	h.voiceDebug = debug
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

	tokenMask := token
	if len(token) > 8 {
		tokenMask = token[:4] + "..." + token[len(token)-4:]
	}
	slog.Info("validating invite", "token", tokenMask)
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
		Turn:        sess.TurnNumber,
		HasHistory:  len(sess.ConversationHistory) > 0,
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
		turnAccMu          sync.Mutex
		committedPhrases   []string
		activeInterim      string
		lastSpeechActivity time.Time
		speechActive       bool
		lastCommittedTurn  int64 = -1
		turnCommitMu       sync.Mutex
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
				Turn: int(curTurn.Load()),
			})
		}
		flushAudioBuffer(int(targetTurn))
		turnWg.Add(1)
		finalTranscriptChan <- clean
		return true
	}

	// Goroutine B — STT Session Manager with 2200ms Conversational Silence Debounce
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
				// Do not trigger turn completion while Voxi is actively speaking
				if tc.isSpeaking() {
					continue
				}

				turnAccMu.Lock()
				var parts []string
				for _, p := range committedPhrases {
					parts = append(parts, p)
				}
				if activeInterim != "" {
					parts = append(parts, activeInterim)
				}
				fullCandidateText := strings.TrimSpace(strings.Join(parts, " "))
				hasSpoken := speechActive && len(strings.Fields(fullCandidateText)) >= 2 && !isFillerOnly(fullCandidateText)
				silenceDuration := time.Since(lastSpeechActivity)

				// Require 2200ms (2.2 seconds) of sustained silence before considering candidate's turn complete
				shouldAutoCommit := hasSpoken && silenceDuration >= 2200*time.Millisecond
				if shouldAutoCommit {
					committedPhrases = nil
					activeInterim = ""
					speechActive = false
				}
				turnAccMu.Unlock()

				if shouldAutoCommit {
					target := curTurn.Load() + 1
					if commitTurn(fullCandidateText, target) {
						slog.Info("conversational silence wait committed speech turn",
							"interview_id", interviewID, "turn", target,
							"silence_ms", silenceDuration.Milliseconds(), "text", fullCandidateText)
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
							turnAccMu.Lock()
							if clean != "" {
								speechActive = true
								lastSpeechActivity = time.Now()
								if res.IsFinal {
									committedPhrases = append(committedPhrases, clean)
									activeInterim = ""
								} else {
									activeInterim = clean
								}
							} else if res.SpeechFinal {
								lastSpeechActivity = time.Now()
							}

							// Accumulate full transcript for the current candidate turn
							var parts []string
							for _, p := range committedPhrases {
								parts = append(parts, p)
							}
							if activeInterim != "" {
								parts = append(parts, activeInterim)
							}
							fullText := strings.TrimSpace(strings.Join(parts, " "))
							turnAccMu.Unlock()

							if fullText != "" {
								send(WSMessage{
									Type:    "interim_transcript",
									Text:    fullText,
									IsFinal: res.IsFinal,
								})
								if clean != "" {
									if tc.isSpeaking() && !isFillerOnly(clean) {
										tc.interrupt()
										send(WSMessage{
											Type: "ai_interrupted",
											Turn: int(curTurn.Load()),
										})
									}
									if h.ragPrefetcher != nil {
										h.ragPrefetcher.TriggerPrefetch(ctx, interviewID, docketID,
											int(curTurn.Load())+1, fullText)
									}
									if h.voiceDebug {
										slog.Info("voice_hop: stt_interim", "turn", curTurn.Load()+1, "text", fullText)
									}
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

	// 4. Initial Greeting on new or unstarted session
	initialGreetingDone := make(chan struct{})
	if len(sess.ConversationHistory) == 0 {
		sess.TurnNumber = 1
		curTurn.Store(1)
		roleTitle := sess.Docket.Title
		if roleTitle == "" {
			roleTitle = "Technical Assessment"
		}
		initialPrompt := fmt.Sprintf("Candidate has entered the stage for the %s assessment. Introduce yourself as Voxi, explain the interview purpose and conversational format, and invite the candidate to introduce themselves.", roleTitle)
		initialHistory := []model.Message{
			{Role: "participant", Content: initialPrompt},
		}

		turnWg.Add(1)
		go func() {
			defer turnWg.Done()
			defer close(initialGreetingDone)
			turnCtx, turnCancel := tc.startTurn()
			defer turnCancel()
			defer tc.done()
			if err := h.sendAITurn(turnCtx, send, sess, 1, buildIntroSystemPrompt(&sess.Docket), initialHistory); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("initial assessor turn failed", "interview_id", interviewID, "error", err)
				sendError("INTERNAL_ERROR", "Internal server error")
			}
		}()
	} else {
		close(initialGreetingDone)
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

				// Sequence barrier: Ensure initial greeting finishes saving/publishing before participant turn 2
				select {
				case <-connDone:
					turnWg.Done()
					return
				case <-initialGreetingDone:
				}
				if isFatalErr() {
					turnWg.Done()
					return
				}

				sess.TurnNumber++
				turn := sess.TurnNumber
				curTurn.Store(int64(turn))

				// RAG lookup runs in parallel with the coverage update; it is served
				// from the interim-transcript prefetch cache when available.
				ragCh := make(chan string, 1)
				go func(finalText string, n int) {
					if h.ragPrefetcher != nil {
						ragCtx, _ := h.ragPrefetcher.GetContext(ctx, interviewID, docketID, n, finalText)
						ragCh <- ragCtx
					} else {
						ragCh <- ""
					}
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

		case "finish_turn", "candidate_done_speaking":
			if !tc.isSpeaking() {
				turnAccMu.Lock()
				var parts []string
				for _, p := range committedPhrases {
					parts = append(parts, p)
				}
				if activeInterim != "" {
					parts = append(parts, activeInterim)
				}
				fullCandidateText := strings.TrimSpace(strings.Join(parts, " "))
				hasSpoken := len(strings.Fields(fullCandidateText)) >= 1 && !isFillerOnly(fullCandidateText)
				committedPhrases = nil
				activeInterim = ""
				speechActive = false
				turnAccMu.Unlock()

				if hasSpoken {
					target := curTurn.Load() + 1
					if commitTurn(fullCandidateText, target) {
						slog.Info("manual finish_turn committed speech turn",
							"interview_id", interviewID, "turn", target, "text", fullCandidateText)
					}
				}
			}

		case "reset_session":
			slog.Info("reset_session requested", "interview_id", interviewID)
			tc.interrupt()
			turnCommitMu.Lock()
			lastCommittedTurn = 0
			turnCommitMu.Unlock()
			turnAccMu.Lock()
			committedPhrases = nil
			activeInterim = ""
			speechActive = false
			turnAccMu.Unlock()
			turnWg.Wait()
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
			roleTitle := sess.Docket.Title
			if roleTitle == "" {
				roleTitle = "Technical Assessment"
			}
			initialPrompt := fmt.Sprintf("Candidate has entered the stage for the %s assessment. Introduce yourself as Voxi, explain the interview purpose and conversational format, and invite the candidate to introduce themselves.", roleTitle)
			initialHistory := []model.Message{
				{Role: "participant", Content: initialPrompt},
			}
			turnWg.Add(1)
			go func() {
				defer turnWg.Done()
				turnCtx, turnCancel := tc.startTurn()
				defer turnCancel()
				defer tc.done()
				if err := h.sendAITurn(turnCtx, send, sess, 1, buildIntroSystemPrompt(&sess.Docket), initialHistory); err != nil && !errors.Is(err, context.Canceled) {
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

	if h.voiceDebug {
		slog.Info("voice_hop: llm_start", "turn", turnNum)
	}

	streamChan := make(chan string, 100)
	llmErrChan := make(chan error, 1)
	go func() {
		defer close(streamChan)
		err := h.llmPool.GenerateResponse(ctx, systemPrompt, history, streamChan)
		if err != nil && ctx.Err() == nil {
			slog.Error("llm streaming response failed", "turn", turnNum, "error", err)
			llmErrChan <- err
		}
	}()

	var fullText strings.Builder
	var sentenceBuf strings.Builder

	var deliveredMu sync.Mutex
	var deliveredText strings.Builder

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
					deliveredMu.Lock()
					if deliveredText.Len() > 0 {
						deliveredText.WriteString(" ")
					}
					deliveredText.WriteString(trimmed)
					deliveredMu.Unlock()

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

			deliveredMu.Lock()
			spoken := strings.TrimSpace(deliveredText.String())
			deliveredMu.Unlock()

			if spoken != "" {
				sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
					Role:      "assessor",
					Content:   spoken + " [interrupted]",
					Timestamp: time.Now().UTC(),
				})
			} else if fullText.Len() > 0 {
				sess.ConversationHistory = append(sess.ConversationHistory, model.Message{
					Role:      "assessor",
					Content:   "[interrupted before speech]",
					Timestamp: time.Now().UTC(),
				})
			}
			if len(sess.ConversationHistory) > 0 {
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
		select {
		case err := <-llmErrChan:
			return fmt.Errorf("llm generation failed: %w", err)
		default:
			return errors.New("empty LLM response")
		}
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

// buildIntroSystemPrompt generates the system prompt for the orientation/introductory turn (Turn 1).
func buildIntroSystemPrompt(docket *model.Docket) string {
	roleTitle := docket.Title
	if roleTitle == "" {
		roleTitle = "Technical Assessment"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are Voxi, an advanced autonomous technical voice interviewer conducting an assessment for the %q position.\n\n", roleTitle))
	sb.WriteString("YOUR OBJECTIVE FOR THIS INITIAL TURN:\n" +
		"1. Warmly welcome the candidate and introduce yourself as Voxi.\n" +
		fmt.Sprintf("2. Explain that today's session is an assessment for the %q role, designed to explore their hands-on architecture, engineering decisions, and practical problem-solving.\n", roleTitle) +
		"3. Explain the pattern of the interview: this is an interactive, conversational technical dialogue where questions adapt dynamically; they can speak naturally, take their time, and interrupt you at any point.\n" +
		"4. Conclude by inviting the candidate to introduce themselves: ask them to share a brief background about who they are and what technical projects or systems they have been building recently.\n" +
		"5. STRICT RULE: Do NOT ask any technical questions or rubric questions yet in this turn. Focus solely on introduction, setting expectations, and inviting their self-introduction.\n\n" +
		"CRITICAL VOICE & BREVITY CONSTRAINTS:\n" +
		"- Never use markdown formatting in your responses. No asterisks, no bold, no bullet points, no hashtags, no backticks. Respond in plain conversational spoken sentences only.\n" +
		"- Speak warmly, clearly, and concisely in 3 to 4 spoken sentences.\n")
	return sb.String()
}

// buildSystemPrompt renders the dynamic project-centric assessment prompt leveraging
// candidate projects, docket instructions, rubric criteria, and adaptive directives.
func buildSystemPrompt(docket *model.Docket, history []model.Message,
	focusName, focusDesc, ragContext string) string {
	var sb strings.Builder
	roleTitle := docket.Title
	if roleTitle == "" {
		roleTitle = "Technical Assessment"
	}
	sb.WriteString(fmt.Sprintf("You are Voxi, an elite autonomous technical voice interviewer conducting an in-depth assessment for the %q position.\n\n", roleTitle))

	sb.WriteString("CORE INTERVIEW DIRECTIVE & CANDIDATE PROJECT CENTRICITY:\n" +
		"- GROUND QUESTIONS IN THE CANDIDATE'S PROJECTS: The candidate's self-introduction and the specific systems or projects they describe are the primary technical basis for this assessment. Listen attentively to their background.\n" +
		"- DYNAMICALLY ADAPT: Do NOT read question prompts like an inflexible textbook robot. Instead, frame your technical inquiries around the candidate's actual projects and architectures (for example, connect the topic of caching, message queues, concurrency, or consensus to the projects the candidate built).\n" +
		"- DIALECTIC PROBING: When the candidate answers, follow up on their specific engineering trade-offs ('Why did you choose that approach?', 'What failure modes did you consider?', 'How did it perform under peak load?').\n" +
		"- NEVER ANSWER YOUR OWN QUESTIONS: You are the interviewer, NOT the candidate or an instructor. Never explain the answers or lecture the candidate.\n\n")

	if len(history) <= 2 {
		sb.WriteString("STAGE NOTE (Turn 2 - First Technical Inquiry):\n" +
			"The candidate has just introduced themselves and their projects. Warmly acknowledge the specific projects or technologies they shared in 1 short sentence, and then ask your first technical question directly probing the architecture, trade-offs, or implementation challenges of their projects, aligning with the assessment blueprint.\n\n")
	} else {
		sb.WriteString("STAGE NOTE (Follow-up Turn):\n" +
			"Acknowledge the candidate's last answer and follow up on the engineering trade-offs, edge cases, or transition smoothly to assess the next competency in their project context.\n\n")
	}

	if docket.Instructions != "" {
		sb.WriteString("Hiring Manager Instructions & Focus Guidelines:\n")
		sb.WriteString(docket.Instructions)
		sb.WriteString("\n\n")
	}

	if len(docket.Rubric) > 0 {
		sb.WriteString("Core Evaluation Criteria (Competencies to Probe):\n")
		for _, r := range docket.Rubric {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", r.Name, r.Description))
		}
		sb.WriteString("\n")
	}

	if len(docket.Questions) > 0 {
		sb.WriteString("Blueprint Inquiry Themes & Core Questions (Use as thematic guides tailored to the candidate's projects, NOT a rigid script):\n")
		for i, q := range docket.Questions {
			hint := ""
			if q.FollowUpHint != nil && *q.FollowUpHint != "" {
				hint = fmt.Sprintf(" (Probe angle: %s)", *q.FollowUpHint)
			}
			sb.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, q.QuestionText, hint))
		}
		sb.WriteString("\n")
	}

	if ragContext != "" {
		sb.WriteString("Retrieved Knowledge Base Context:\n")
		sb.WriteString(ragContext)
		sb.WriteString("\n\n")
	}

	if focusName != "" {
		sb.WriteString(fmt.Sprintf(
			"CURRENT ADAPTIVE FOCUS: The candidate has not yet demonstrated sufficient depth in [%s] (%s). Your next question MUST probe this competency within their project context.\n\n",
			focusName, focusDesc))
	}

	sb.WriteString("CRITICAL VOICE & BREVITY CONSTRAINTS:\n" +
		"- Never use markdown formatting in your responses. No asterisks, no bold, no bullet points, no hashtags, no backticks. Respond in plain conversational spoken sentences only.\n" +
		"- BREVITY IS MANDATORY: Speak strictly 2 to 3 concise sentences (under 45 words total). Keep questions clear, focused, and conversational so the candidate has the floor.\n")
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

