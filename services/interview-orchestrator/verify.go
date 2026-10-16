package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

type WSMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
	Message string `json:"message,omitempty"`
}

func main() {
	ctx := context.Background()
	// 1. Setup DB fixtures
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/interview_platform?sslmode=disable"
	}
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer conn.Close(ctx)

	orgID := uuid.New()
	userID := uuid.New()
	candidateID := uuid.New()
	jobRoleID := uuid.New()
	docketID := uuid.New()

	_, err = conn.Exec(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, 'Test Org', $2)`, orgID, "test-org-"+orgID.String())
	if err != nil {
		log.Fatalf("insert org: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO users (id, org_id, email, password_hash, full_name) VALUES ($1, $2, $3, 'hash', 'Test User')`, userID, orgID, userID.String()+"@example.com")
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO assessment_tracks (id, org_id, created_by, title, description) VALUES ($1, $2, $3, 'Test Job', 'Desc')`, jobRoleID, orgID, userID)
	if err != nil {
		log.Fatalf("insert job: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO participants (id, org_id, job_role_id, full_name, email) VALUES ($1, $2, $3, 'Test Candidate', 'cand@example.com')`, candidateID, orgID, jobRoleID)
	if err != nil {
		log.Fatalf("insert candidate: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO assessment_blueprints (id, org_id, job_role_id, title, instructions) VALUES ($1, $2, $3, 'Test Docket', 'Instructions')`, docketID, orgID, jobRoleID)
	if err != nil {
		log.Fatalf("insert docket: %v", err)
	}

	// 2. Generate an auth token for JCS
	// Actually, we can bypass Auth for CreateInvite if we just simulate a JWT. JCS uses JWT secret.
	// But it's easier to create a proper login. Let's do that.
	// No, CreateInvite requires JWT middleware.
	// Wait, we can just insert the Invite directly into the DB for the test!
	inviteToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	inviteID := uuid.New()
	interviewID := uuid.New()

	_, err = conn.Exec(ctx, `
		INSERT INTO assessment_invites (id, candidate_id, docket_id, org_id, token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		inviteID, candidateID, docketID, orgID, inviteToken, expiresAt)
	if err != nil {
		log.Fatalf("insert invite: %v", err)
	}
	_, err = conn.Exec(ctx, `
		INSERT INTO assessment_sessions (id, invite_id, candidate_id, docket_id, org_id, status)
		VALUES ($1, $2, $3, $4, $5, 'invited')`,
		interviewID, inviteID, candidateID, docketID, orgID)
	if err != nil {
		log.Fatalf("insert interview: %v", err)
	}

	fmt.Println("Fixtures created. Invite Token:", inviteToken)

	// 3. Connect WS
	wsURL := fmt.Sprintf("ws://localhost:8002/ws/interview/?token=%s", inviteToken)
	c1, resp1, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp1 != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp1.Body)
			log.Fatalf("WS connect c1 failed: %v, body: %s", err, buf.String())
		}
		log.Fatalf("WS connect c1 failed: %v", err)
	}
	defer c1.Close()

	fmt.Println("c1 connected")

	var msg WSMessage
	if err := c1.ReadJSON(&msg); err != nil {
		log.Fatalf("c1 read error: %v", err)
	}
	if msg.Type != "connected" {
		b, _ := json.Marshal(msg)
		log.Fatalf("c1 expected connected, got %s, payload: %s", msg.Type, string(b))
	}
	fmt.Println("c1 received 'connected'")

	// 4. Test Distributed Lock (Task 6)
	c2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("WS connect c2 failed: %v", err)
	}
	defer c2.Close()

	if err := c2.ReadJSON(&msg); err != nil {
		log.Fatalf("c2 read error: %v", err)
	}
	if msg.Type != "error" || msg.Message != "Interview already in progress" {
		log.Fatalf("c2 expected error 'Interview already in progress', got %s: %s", msg.Type, msg.Message)
	}
	fmt.Println("c2 received expected lock error")

	// 5. Test End-to-End Reconnect (Task 5)
	c1.Close()
	fmt.Println("c1 disconnected")
	time.Sleep(2 * time.Second) // allow session manager to release lock

	c3, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("WS connect c3 failed: %v", err)
	}
	defer c3.Close()

	if err := c3.ReadJSON(&msg); err != nil {
		log.Fatalf("c3 read error: %v", err)
	}
	if msg.Type != "connected" {
		log.Fatalf("c3 expected connected, got %s", msg.Type)
	}
	fmt.Println("c3 reconnected successfully")

	// 6. Test Interview Completion Idempotency (Task 8)
	err = c3.WriteJSON(WSMessage{Type: "end_interview"})
	if err != nil {
		log.Fatalf("c3 write error: %v", err)
	}
	
	// Expect interview_complete
	for {
		if err := c3.ReadJSON(&msg); err != nil {
			log.Fatalf("c3 read loop error: %v", err)
		}
		if msg.Type == "interview_complete" {
			fmt.Println("c3 received 'interview_complete'")
			break
		}
	}
	
	c3.Close()

	// Connect a fourth time to trigger completion again, it should fail validation because status=completed
	c4, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		if err := c4.ReadJSON(&msg); err == nil {
			fmt.Printf("c4 unexpected message: %+v\n", msg)
			if msg.Type != "error" {
				log.Fatalf("c4 expected connection error due to completed status, but connected")
			}
		}
		c4.Close()
	} else {
		fmt.Printf("c4 dial failed as expected: %v\n", err)
	}

	fmt.Println("All integration tests passed successfully.")
}
