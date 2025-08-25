package main

import (
	"flag"
	"fmt"
	"slices"
	_ "time/tzdata"

	"github.com/BurntSushi/toml"
)

type Config struct {
	LogLevel int    `toml:"log_level"`
	CronJob  string `toml:"cron_job"`
	// Deprecated. UseCron is deprecated, if the CronJob field is empty,
	// This well known run as CronJob.
	UseCron  bool               `toml:"use_cron"`
	WaitTime int                `toml:"wait_time"` // based on second.
	Clients  map[string]Clients `toml:"clients"`
	SMTP     map[string]SMTP    `toml:"smtp"`
}

type SMTP struct {
	Mail       string         `toml:"mail"`
	Address    string         `toml:"host"`
	Port       string         `toml:"port"`
	From       string         `toml:"from"`
	Username   string         `toml:"username"`
	Password   string         `toml:"password"`
	AuthMethod smtpAuthMethod `toml:"auth_method"`
	Identity   string         `toml:"identity"`
	SkipTLS    bool           `toml:"skip_tls"`
}

type Clients struct {
	Smtp       string   `toml:"smtp"`
	BillID     string   `toml:"bill_id"`
	BillIDs    []string `toml:"bill_ids"`
	AuthToken  string   `toml:"auth_token"`
	Recipients []string `toml:"recipients"`
}

type smtpAuthMethod string

const (
	smtpAuthMethodPlain  smtpAuthMethod = "plain"
	smtpAuthMethodMD5    smtpAuthMethod = "cram-md5"
	smtpAuthMethodCustom smtpAuthMethod = "custom"
)

var smtpAuthMethodValues = []smtpAuthMethod{smtpAuthMethodPlain, smtpAuthMethodMD5, smtpAuthMethodCustom}

func ParseConfig() (*Config, error) {
	var configFilePath string
	flag.StringVar(&configFilePath, "file", "config.toml", "config file(toml formatted)")
	flag.Parse()

	config := new(Config)
	if _, err := toml.DecodeFile(configFilePath, config); err != nil {
		return nil, err
	}

	for _, smtp := range config.SMTP {
		if !slices.Contains(smtpAuthMethodValues, smtp.AuthMethod) {
			return nil, fmt.Errorf("invalid smtp auth, should be exactly one of %v", smtpAuthMethodValues)
		}
	}

	return config, nil
}
