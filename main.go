package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

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
		for _, client := range config.Clients {
			data, err := PlannedBlackOut(context.Background(), client.AuthToken, client.BillID, time.Now().AddDate(0, 0, -7), time.Now().AddDate(0, 0, -7))
			if err != nil {
				slog.Error("plannedBlackOut failed", "error", err)
				continue
			}

			if err := mail.Do(data, client.Recipients); err != nil {
				slog.Error("failed to send mail", "error", err)
				continue
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
