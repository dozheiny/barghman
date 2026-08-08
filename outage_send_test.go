package main_test

import (
	"testing"
	"time"

	main "github.com/dozheiny/barghman"
	"github.com/stretchr/testify/require"
)

func TestDiffRecipients(t *testing.T) {
	t.Parallel()

	cached := []string{"a@example.com", "b@example.com"}
	current := []string{"b@example.com", "c@example.com", "d@example.com"}

	got := main.DiffRecipients(cached, current)
	require.Equal(t, []string{"c@example.com", "d@example.com"}, got)
}

func TestDiffRecipientsNone(t *testing.T) {
	t.Parallel()

	cached := []string{"a@example.com", "b@example.com"}
	current := []string{"a@example.com", "b@example.com"}

	require.Empty(t, main.DiffRecipients(cached, current))
}

func TestDecideOutageSendNew(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	recipients := []string{"a@example.com"}

	d := main.DecideOutageSend(nil, false, start, end, recipients)
	require.Equal(t, main.OutageSendNew, d.Action)
	require.Equal(t, recipients, d.Recipients)
	require.Equal(t, uint(0), d.Sequence)
}

func TestDecideOutageSendSkipUnchanged(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	cached := &main.FileContent{
		Sequence:            2,
		StartOutageDateTime: start,
		EndOutageDateTime:   end,
		Recipients:          []string{"a@example.com"},
	}

	d := main.DecideOutageSend(cached, true, start, end, []string{"a@example.com"})
	require.Equal(t, main.OutageSendSkip, d.Action)
}

func TestDecideOutageSendEndTimeOnlyUpdate(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	newEnd := start.Add(3 * time.Hour)
	recipients := []string{"a@example.com", "b@example.com"}

	cached := &main.FileContent{
		Sequence:            1,
		StartOutageDateTime: start,
		EndOutageDateTime:   end,
		Recipients:          []string{"a@example.com"},
	}

	d := main.DecideOutageSend(cached, true, start, newEnd, recipients)
	require.Equal(t, main.OutageSendUpdate, d.Action)
	require.Equal(t, recipients, d.Recipients)
	require.Equal(t, uint(2), d.Sequence)
}

func TestDecideOutageSendNewRecipientsOnly(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	cached := &main.FileContent{
		Sequence:            3,
		StartOutageDateTime: start,
		EndOutageDateTime:   end,
		Recipients:          []string{"a@example.com"},
	}

	d := main.DecideOutageSend(cached, true, start, end, []string{"a@example.com", "b@example.com"})
	require.Equal(t, main.OutageSendNewRecipients, d.Action)
	require.Equal(t, []string{"b@example.com"}, d.Recipients)
	require.Equal(t, uint(3), d.Sequence)
}

func TestDecideOutageSendUpdateCoversNewRecipients(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	newStart := start.Add(30 * time.Minute)
	recipients := []string{"a@example.com", "b@example.com"}

	cached := &main.FileContent{
		Sequence:            0,
		StartOutageDateTime: start,
		EndOutageDateTime:   end,
		Recipients:          []string{"a@example.com"},
	}

	d := main.DecideOutageSend(cached, true, newStart, end, recipients)
	require.Equal(t, main.OutageSendUpdate, d.Action)
	require.Equal(t, recipients, d.Recipients)
	require.Equal(t, uint(1), d.Sequence)
}
