package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

const appName = "barghman"

var cachePathDir string

func init() {
	cachePath, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("Unable to get user cache directory: %v", err)
	}

	cachePathDir = cachePath + "/" + appName + "/"

	if err := os.MkdirAll(cachePathDir, 0o755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatalf("Failed to create cache path directory: %v", err)
		}
	}
}

func main() {
	config, err := ParseConfig()
	if err != nil {
		slog.Error("Failed to parse config", "error", err)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(slog.Level(config.LogLevel))
	slog.Debug("config file loaded", "config", config)

	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		slog.Error("Unable to load location", "error", err)
		os.Exit(1)
	}

	var smtp SMTP
	for _, s := range config.SMTP {
		smtp = s
	}

	mail, err := NewMailClient(smtp, location)
	if err != nil {
		slog.Error("Unable to create new mail client", "error", err)
		os.Exit(1)
	}

	job := func() {
		slog.Debug("job started")

		for subject, c := range config.Clients {
			data, err := PlannedBlackOut(context.Background(), c.AuthToken, c.BillID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 5))
			if err != nil {
				slog.Error("plannedBlackOut failed", "error", err)
				continue
			}

			for _, d := range data {
				startDate, endDate, err := d.ParseTime(location)
				if err != nil {
					slog.Error("Failed to parse time", "error", err)
					continue
				}

				f, err := LoadOrCreateFile(c.BillID, d.OutageNumber, startDate)
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

				fcf, err = d.ToFileContent(location, c.BillID, c.Recipients, sequence)
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

		}
	}

	if !config.UseCron {
		job()
		return
	}

	c := cron.New(cron.WithLocation(location))
	if _, err := c.AddFunc(config.CronJob, job); err != nil {
		slog.Error("couldn't add to the cron job", "error", err)
		os.Exit(1)
	}

	defer c.Stop()
	c.Start()

	select {}
}
