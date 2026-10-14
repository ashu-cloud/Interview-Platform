package model

import "testing"

const (
	sid = "11111111-1111-1111-1111-111111111111"
	oid = "22222222-2222-2222-2222-222222222222"
)

func TestParseTurnEventLegacyRuntimePayload(t *testing.T) {
	ev, err := ParseTurnEvent([]byte(`{"event_type":"turn","interview_id":"` + sid + `","org_id":"` + oid +
		`","turn_number":2,"role":"assistant","content":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.SessionID != sid || ev.Role != RoleAssessor || ev.TurnNumber != 2 {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestParseTurnEventSessionShape(t *testing.T) {
	ev, err := ParseTurnEvent([]byte(`{"session_id":"` + sid + `","org_id":"` + oid +
		`","turn_number":1,"role":"participant","content":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.SessionID != sid || ev.Role != RoleParticipant {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestParseTurnEventRejectsMalformed(t *testing.T) {
	bad := map[string]string{
		"not json":     `{`,
		"bad session":  `{"session_id":"x","org_id":"` + oid + `","turn_number":1,"role":"user","content":"a"}`,
		"bad org":      `{"session_id":"` + sid + `","org_id":"x","turn_number":1,"role":"user","content":"a"}`,
		"bad role":     `{"session_id":"` + sid + `","org_id":"` + oid + `","turn_number":1,"role":"bot","content":"a"}`,
		"empty text":   `{"session_id":"` + sid + `","org_id":"` + oid + `","turn_number":1,"role":"user","content":""}`,
		"negative num": `{"session_id":"` + sid + `","org_id":"` + oid + `","turn_number":-1,"role":"user","content":"a"}`,
	}
	for name, payload := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseTurnEvent([]byte(payload)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
