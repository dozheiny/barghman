package main

import (
	"log/slog"
	"os"
	"time"

	_ "time/tzdata"

	"github.com/robfig/cron/v3"
)

const appName = "barghman"

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

	cachePathDir, err := CreateCachePath()
	if err != nil {
		slog.Error("failed to create cache path", "error", err)
		os.Exit(1)
	}

	jobFunc := MailerFunc(cachePathDir, *config, location)
	deleteFunc := DeleteCacheFunc(cachePathDir, config.DeleteDurationPeriod)

	if len(config.CronJob) == 0 {
		jobFunc()
		return
	}

	c := cron.New(cron.WithLocation(location))

	if _, err := c.AddFunc(config.CronJob, jobFunc); err != nil {
		slog.Error("couldn't add mailer func to the cron job", "error", err)
		os.Exit(1)
	}

	if _, err := c.AddFunc("@daily", deleteFunc); err != nil {
		slog.Error("couldn't add delete func to the cron job", "error", err)
		os.Exit(1)
	}

	defer c.Stop()
	c.Start()

	select {}
}
