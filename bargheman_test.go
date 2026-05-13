package main_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	main "github.com/dozheiny/barghman"
	"github.com/stretchr/testify/require"
)

func TestOneDayDifferentHours(t *testing.T) {
	loc := time.Local

	body, err := os.ReadFile("test_data/one_day_different_hours.json")
	require.NoError(t, err)
	require.NotNil(t, body)

	pbor := new(main.PlannedBlackOutResponse)
	require.NoError(t, json.Unmarshal(body, pbor))

	var sequence uint
	for _, d := range pbor.Data {
		startDate, endDate, err := d.ParseTime(loc)
		require.NoError(t, err)

		f, err := main.LoadOrCreateFile(t.TempDir(), strconv.Itoa(d.OutageNumber), d.OutageNumber, startDate)
		require.NoError(t, err)

		defer f.Close()

		var fileData []byte
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fileData = append(fileData, scanner.Bytes()...)
		}

		require.NoError(t, scanner.Err())

		fcf := new(main.FileContent)

		if len(fileData) != 0 {
			require.NoError(t, json.Unmarshal(fileData, fcf))

			// Checks that the file loaded the start and end datetime is changed or not.
			// If it doesn't changes, ignore it; If it changes, update it.
			if fcf.StartOutageDateTime.Equal(startDate) || fcf.EndOutageDateTime.Equal(endDate) {
				t.Log("This data is already sent as email", "file name", fcf.FileName())
				continue
			}

			sequence = fcf.Sequence + 1
		}

		fcf, err = d.ToFileContent(loc, "", []string{}, sequence)
		require.NoError(t, err)

		require.NoError(t, fcf.Write(f))
	}

	require.Equal(t, uint(1), sequence)
}

func TestDeleteCacheFunc(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files: one old, one new
	oldFile := filepath.Join(tmpDir, "old.cache")
	newFile := filepath.Join(tmpDir, "new.cache")

	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	// Set the modtime of the old file to 48h ago
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old file time: %v", err)
	}

	// Run the cache cleaner with a 24h cutoff
	cleaner := main.DeleteCacheFunc(tmpDir, 24*time.Hour)
	cleaner()

	// Assert old file is deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("expected old file to be deleted, but it exists")
	}

	// Assert new file is still present
	if _, err := os.Stat(newFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("expected new file to remain, but it was deleted")
		} else {
			t.Errorf("unexpected error checking new file: %v", err)
		}
	}
}
