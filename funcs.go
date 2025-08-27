package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

func DeleteCacheFunc(cachePathDir string, period time.Duration) func() {
	return func() {
		u := time.Now().Add(-period)

		files, err := os.ReadDir(cachePathDir)
		if err != nil {
			slog.Error("couldn't read all directories", "error", err)
			return
		}

		for _, f := range files {
			info, err := f.Info()
			if err != nil {
				slog.Error("couldn't read info files", "error", err, "file name", f.Name())
				continue
			}

			if info.ModTime().Before(u) {
				filePath := cachePathDir + "/" + info.Name()
				slog.Debug("removing cache", "file name", filePath)

				if err := os.Remove(filePath); err != nil {
					slog.Error("cannot remove the file", "error", err, "file path", filePath)
					continue
				}
			}
		}
	}
}

func MailerFunc(cachePathDir string, config Config, location *time.Location) func() {
	return func() {
		slog.Debug("job started")

		for subject, c := range config.Clients {

			smtp, ok := config.SMTP[c.SMTP]
			if !ok {
				slog.Error("Cannot map between smtp config and client config", "smtp name", c.SMTP)
				continue
			}

			mail := NewMailClient(smtp, location)

			for _, billID := range append(c.BillIDs, c.BillID) {
				data, err := PlannedBlackOut(context.Background(), c.AuthToken, billID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 5))
				if err != nil {
					slog.Error("PlannedBlackOut failed", "error", err)
					continue
				}

				for _, d := range data {
					startDate, endDate, err := d.ParseTime(location)
					if err != nil {
						slog.Error("Failed to parse time", "error", err)
						continue
					}

					f, err := LoadOrCreateFile(cachePathDir, billID, d.OutageNumber, startDate)
					if err != nil {
						slog.Error("couldn't load or create file", "error", err)
						continue
					}

					defer f.Close()

					var fileData []byte
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						fileData = append(fileData, scanner.Bytes()...)
					}

					if err := scanner.Err(); err != nil {
						slog.Error("scanner returns error", "error", err)
						continue
					}

					fcf := new(FileContent)
					var sequence uint

					if len(fileData) != 0 {
						if err := json.Unmarshal(fileData, fcf); err != nil {
							slog.Error("decode the file data failed", "error", err)
							continue
						}

						// Checks that the file loaded the start and end datetime is changed or not.
						// If it doesn't changes, ignore it; If it changes, update it.
						if fcf.StartOutageDateTime.Equal(startDate) || fcf.EndOutageDateTime.Equal(endDate) {
							slog.Info("This data is already sent as email", "file name", fcf.FileName())
							continue
						}

						sequence = fcf.Sequence + 1
					}

					fcf, err = d.ToFileContent(location, billID, c.Recipients, sequence)
					if err != nil {
						slog.Error("Failed to convert data to file content", "error", err)
						continue
					}

					if err := mail.Do(fcf, subject); err != nil {
						slog.Error("Failed to send mail", "error", err)
						continue
					}

					if err := fcf.Write(f); err != nil {
						slog.Error("Failed to cache data", "error", err)
					}
				}

				time.Sleep(time.Second * time.Duration(config.WaitTime))
			}
		}

		slog.Debug("all clients sent, waiting for next cron cycle")
	}
}
