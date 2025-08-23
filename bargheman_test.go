package main_test

import (
	"bufio"
	"encoding/json"
	"os"
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

		f, err := main.LoadOrCreateFile(strconv.Itoa(d.OutageNumber), d.OutageNumber, startDate)
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
