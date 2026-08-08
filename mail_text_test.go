package main

import (
	"strings"
	"testing"
	"time"
)

func TestPlainTextPartIncludesWhatWhyHow(t *testing.T) {
	t.Parallel()

	fc := &FileContent{
		Address:             "Test St",
		FarsiOutageDate:     "1405/05/17",
		StartOutageDateTime: time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC),
		EndOutageDateTime:   time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC),
		ReasonOutage:        "maintenance",
	}

	body := plainTextPart("bound", fc, MailKindNew)
	for _, want := range []string{
		"Content-Type: text/plain",
		"planned power outage",
		"listed as a recipient",
		"Address: Test St",
		"Reason: maintenance",
		"How to add this to your calendar",
		"invite.ics",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("plain text missing %q\nbody:\n%s", want, body)
		}
	}

	update := plainTextPart("bound", fc, MailKindUpdate)
	if !strings.Contains(update, "schedule has changed") {
		t.Fatalf("update body missing change notice:\n%s", update)
	}
}
