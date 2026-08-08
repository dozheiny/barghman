package main

import (
	"context"
	"encoding/json"
	"io"
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

func MailerFunc(cachePathDir, configPath string, location *time.Location) func() {
	return func() {
		slog.Debug("job started")

		config, err := LoadConfig(configPath)
		if err != nil {
			slog.Error("Failed to reload config", "error", err, "path", configPath)
			return
		}
		slog.SetLogLoggerLevel(slog.Level(config.LogLevel))

		for subject, c := range config.Clients {
			smtp, ok := config.SMTP[c.SMTP]
			if !ok {
				slog.Error("Cannot map between smtp config and client config", "smtp name", c.SMTP)
				continue
			}

			mail := NewMailClient(smtp, location)

			for _, billID := range append(c.BillIDs, c.BillID) {
				data, err := PlannedBlackOut(context.Background(), config.AuthToken, billID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 5))
				if err != nil {
					slog.Error("PlannedBlackOut failed", "error", err)
					continue
				}

				for _, d := range data {
					if err := processOutage(cachePathDir, location, mail, subject, billID, c.Recipients, d); err != nil {
						slog.Error("Failed to process outage", "error", err, "bill_id", billID, "outage_number", d.OutageNumber)
					}
				}

				time.Sleep(time.Second * time.Duration(config.WaitTime))
			}
		}

		slog.Debug("all clients sent, waiting for next cron cycle")
	}
}

func processOutage(cachePathDir string, location *time.Location, mail Mail, subject, billID string, recipients []string, d Data) error {
	startDate, endDate, err := d.ParseTime(location)
	if err != nil {
		slog.Error("Failed to parse time", "error", err)
		return err
	}

	f, err := LoadOrCreateFile(cachePathDir, billID, d.OutageNumber, startDate)
	if err != nil {
		slog.Error("couldn't load or create file", "error", err)
		return err
	}
	defer f.Close()

	cached, hasCache, err := readCacheContent(f)
	if err != nil {
		slog.Error("couldn't read cache file", "error", err)
		return err
	}

	decision := DecideOutageSend(cached, hasCache, startDate, endDate, recipients)
	if decision.Action == OutageSendSkip {
		slog.Info("This data is already sent as email", "file name", FileName(billID, d.OutageNumber, startDate))
		return nil
	}

	fcf, err := d.ToFileContent(location, billID, decision.Recipients, decision.Sequence)
	if err != nil {
		slog.Error("Failed to convert data to file content", "error", err)
		return err
	}

	kind := MailKindNew
	if decision.Action == OutageSendUpdate {
		kind = MailKindUpdate
	}

	if err := mail.Do(fcf, subject, kind); err != nil {
		slog.Error("Failed to send mail", "error", err)
		return err
	}

	// Persist the full current recipient list so removals and additions stick.
	fcf.Recipients = copyStrings(recipients)
	if err := fcf.Write(f); err != nil {
		slog.Error("Failed to cache data", "error", err)
		return err
	}
	return nil
}

func readCacheContent(f *os.File) (*FileContent, bool, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return nil, false, err
	}

	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, false, err
	}
	if len(fileData) == 0 {
		return nil, false, nil
	}

	fcf := new(FileContent)
	if err := json.Unmarshal(fileData, fcf); err != nil {
		return nil, false, err
	}
	return fcf, true, nil
}
