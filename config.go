package main

import (
	"flag"
	_ "time/tzdata"

	"github.com/BurntSushi/toml"
)

type Config struct {
	LogLevel int    `toml:"log_level"`
	CronJob  string `toml:"cron_job"`
	UseCron  bool   `toml:"use_cron"`
	//	CachePathDir string             `toml:"cache_path_dir"`
	Clients map[string]Clients `toml:"clients"`
	SMTP    map[string]SMTP    `toml:"smtp"`
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
	BillID     string   `toml:"bill_id"`
	AuthToken  string   `toml:"auth_token"`
	Recipients []string `toml:"recipients"`
}

type smtpAuthMethod string

const (
	smtpAuthMethodPlain  smtpAuthMethod = "plain"
	smtpAuthMethodMD5    smtpAuthMethod = "cram-md5"
	smtpAuthMethodCustom smtpAuthMethod = "custom"
)

func ParseConfig() (*Config, error) {
	var configFilePath string
	flag.StringVar(&configFilePath, "file", "config.toml", "config file(toml formatted)")
	flag.Parse()

	config := new(Config)
	if _, err := toml.DecodeFile(configFilePath, config); err != nil {
		return nil, err
	}

	return config, nil
}
