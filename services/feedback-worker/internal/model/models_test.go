package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseInterviewCompletedEventCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "canonical",
			payload: `{"session_id":"session-1","org_id":"org-1","participant_id":"participant-1","blueprint_id":"blueprint-1","turn_count":2}`,
		},
		{
			name:    "legacy",
			payload: `{"event_type":"completed","interview_id":"session-1","org_id":"org-1","candidate_id":"participant-1","docket_id":"blueprint-1","turn_count":2}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := ParseInterviewCompletedEvent([]byte(test.payload))
			if err != nil {
				t.Fatal(err)
			}
			if event.InterviewID != "session-1" || event.OrgID != "org-1" ||
				event.CandidateID != "participant-1" || event.DocketID != "blueprint-1" || event.TurnCount != 2 {
				t.Fatalf("unexpected completed event: %+v", event)
			}
		})
	}
}

func TestParseTurnEventCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantRole string
	}{
		{
			name:     "canonical participant",
			payload:  `{"session_id":"session-1","org_id":"org-1","turn_number":1,"role":"participant","content":"answer"}`,
			wantRole: RoleParticipant,
		},
		{
			name:     "canonical assessor",
			payload:  `{"session_id":"session-1","org_id":"org-1","turn_number":1,"role":"assessor","content":"question"}`,
			wantRole: RoleAssessor,
		},
		{
			name:     "legacy user",
			payload:  `{"event_type":"turn","interview_id":"session-1","org_id":"org-1","turn_number":1,"role":"user","content":"answer"}`,
			wantRole: RoleParticipant,
		},
		{
			name:     "legacy assistant",
			payload:  `{"event_type":"turn","interview_id":"session-1","org_id":"org-1","turn_number":1,"role":"assistant","content":"question"}`,
			wantRole: RoleAssessor,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := ParseTurnEvent([]byte(test.payload))
			if err != nil {
				t.Fatal(err)
			}
			if event.InterviewID != "session-1" || event.Role != test.wantRole || event.TurnNumber != 1 {
				t.Fatalf("unexpected turn event: %+v", event)
			}
		})
	}
}

func TestCanonicalEventJSONFields(t *testing.T) {
	completed, err := json.Marshal(InterviewCompletedEvent{
		InterviewID: "session-1",
		CandidateID: "participant-1",
		DocketID:    "blueprint-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	turn, err := json.Marshal(TurnEvent{InterviewID: "session-1", Role: RoleParticipant})
	if err != nil {
		t.Fatal(err)
	}

	for _, payload := range []string{string(completed), string(turn)} {
		if !strings.Contains(payload, `"session_id":"session-1"`) || strings.Contains(payload, `"interview_id"`) {
			t.Fatalf("expected canonical session field in %s", payload)
		}
	}
	if !strings.Contains(string(completed), `"participant_id":"participant-1"`) ||
		!strings.Contains(string(completed), `"blueprint_id":"blueprint-1"`) {
		t.Fatalf("expected canonical completion fields in %s", completed)
	}
}
