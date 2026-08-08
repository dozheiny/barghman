package main

import "time"

// OutageSendAction describes what the mailer should do for a cached outage.
type OutageSendAction int

const (
	OutageSendSkip OutageSendAction = iota
	OutageSendNew
	OutageSendUpdate
	OutageSendNewRecipients
)

// OutageSendDecision is the result of comparing API outage data to cache.
type OutageSendDecision struct {
	Action     OutageSendAction
	Recipients []string
	Sequence   uint
}

// DiffRecipients returns addresses in current that are not in cached,
// preserving the order of current. Matching is case-sensitive.
func DiffRecipients(cached, current []string) []string {
	seen := make(map[string]struct{}, len(cached))
	for _, r := range cached {
		seen[r] = struct{}{}
	}

	var added []string
	for _, r := range current {
		if _, ok := seen[r]; ok {
			continue
		}
		added = append(added, r)
	}
	return added
}

// DecideOutageSend chooses whether to skip, send a new invite, send an
// update for changed times, or notify only newly added recipients.
func DecideOutageSend(cached *FileContent, hasCache bool, start, end time.Time, currentRecipients []string) OutageSendDecision {
	if !hasCache || cached == nil {
		return OutageSendDecision{
			Action:     OutageSendNew,
			Recipients: copyStrings(currentRecipients),
			Sequence:   0,
		}
	}

	timesUnchanged := cached.StartOutageDateTime.Equal(start) && cached.EndOutageDateTime.Equal(end)
	if !timesUnchanged {
		return OutageSendDecision{
			Action:     OutageSendUpdate,
			Recipients: copyStrings(currentRecipients),
			Sequence:   cached.Sequence + 1,
		}
	}

	newRecipients := DiffRecipients(cached.Recipients, currentRecipients)
	if len(newRecipients) == 0 {
		return OutageSendDecision{Action: OutageSendSkip}
	}

	return OutageSendDecision{
		Action:     OutageSendNewRecipients,
		Recipients: newRecipients,
		Sequence:   cached.Sequence,
	}
}

func copyStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
