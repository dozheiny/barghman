package main

import (
	"flag"
	"fmt"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	LogLevel int    `toml:"log_level"`
	CronJob  string `toml:"cron_job"`
	// Deprecated. UseCron is deprecated, if the CronJob field is empty,
	// This well known run as CronJob.
	UseCron bool `toml:"use_cron"`
	// WaitTime is based on second.
	WaitTime int `toml:"wait_time"`
	// DeleteDurationPeriod will be use for delete cache automatically.
	DeleteDurationPeriod time.Duration      `toml:"delete_duration_period"`
	Clients              map[string]Clients `toml:"clients"`
	SMTP                 map[string]SMTP    `toml:"smtp"`

	AuthToken string `toml:"auth_token"`
}

// Proxy holds optional SOCKS5 proxy configuration for outbound SMTP connections.
type Proxy struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// Enabled returns true when a proxy host is configured.
func (p *Proxy) Enabled() bool {
	return p != nil && p.Host != ""
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
	Proxy      *Proxy         `toml:"proxy"`

	// Transport selects how the mail built for this account is delivered.
	// "smtp" (the default) dials the mail server directly on Address:Port.
	// "ews" posts the same MIME message to Exchange Web Services over
	// HTTPS instead, which is useful when SMTP is blocked or throttled but
	// EWS/OWA on 443 is reachable.
	Transport mailTransport `toml:"transport"`
	// EWSURL is the Exchange Web Services endpoint used when Transport is
	// "ews". If empty, it defaults to "https://<host>/EWS/Exchange.asmx".
	EWSURL string `toml:"ews_url"`
}

type Clients struct {
	SMTP       string   `toml:"smtp"`
	BillID     string   `toml:"bill_id"`
	BillIDs    []string `toml:"bill_ids"`
	Recipients []string `toml:"recipients"`
}

type smtpAuthMethod string

const (
	smtpAuthMethodPlain  smtpAuthMethod = "plain"
	smtpAuthMethodMD5    smtpAuthMethod = "cram-md5"
	smtpAuthMethodCustom smtpAuthMethod = "custom"
)

var smtpAuthMethodValues = []smtpAuthMethod{smtpAuthMethodPlain, smtpAuthMethodMD5, smtpAuthMethodCustom}

type mailTransport string

const (
	mailTransportSMTP mailTransport = "smtp"
	mailTransportEWS  mailTransport = "ews"
)

var mailTransportValues = []mailTransport{mailTransportSMTP, mailTransportEWS}

func ParseConfig() (*Config, error) {
	var configFilePath string
	flag.StringVar(&configFilePath, "file", "config.toml", "config file(toml formatted)")
	flag.Parse()

	config := new(Config)
	if _, err := toml.DecodeFile(configFilePath, config); err != nil {
		return nil, err
	}

	for name, smtp := range config.SMTP {
		if smtp.Transport == "" {
			smtp.Transport = mailTransportSMTP
			config.SMTP[name] = smtp
		}

		if !slices.Contains(mailTransportValues, smtp.Transport) {
			return nil, fmt.Errorf("smtp %q: invalid transport, should be exactly one of %v", name, mailTransportValues)
		}

		switch smtp.Transport {
		case mailTransportSMTP:
			if !slices.Contains(smtpAuthMethodValues, smtp.AuthMethod) {
				return nil, fmt.Errorf("smtp %q: invalid smtp auth, should be exactly one of %v", name, smtpAuthMethodValues)
			}
		case mailTransportEWS:
			if smtp.Username == "" || smtp.Password == "" {
				return nil, fmt.Errorf("smtp %q: username and password are required for ews transport", name)
			}
		}
	}

	if config.DeleteDurationPeriod == 0 {
		config.DeleteDurationPeriod = time.Hour * 24 * 7
	}

	return config, nil
}
