# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project overview

Barghman is a Go CLI/service that polls the Iran Power (SAAPA) API for planned
power outages tied to one or more electricity bill IDs, then emails affected
recipients a calendar invite (`.ics`) for each outage. It can run once or on
a cron schedule.

Key files:
- `main.go` - entrypoint, cron wiring.
- `config.go` - TOML config parsing/validation (`Config`, `SMTP`, `Clients`).
- `func.go` - the mailer job (`MailerFunc`) and cache-cleanup job.
- `mail.go` - builds the MIME email (text + `calendar.ics` invite) and sends
  it. Dispatches to either `sendSMTP` (direct SMTP) or `sendEWS`.
- `ews.go` - alternative transport that posts the same MIME message to
  Exchange Web Services over HTTPS (`CreateItem`), for accounts where SMTP
  is blocked/throttled but EWS/OWA is reachable.
- `file.go` - on-disk cache of already-sent outages (`FileContent`), keyed by
  bill ID + outage number + date, so the same outage isn't emailed twice.
- `barghman_client.go` - HTTP client for the SAAPA blackout API.

## Build, test, run

```bash
go build ./...            # build
go vet ./...               # static checks
go test ./...               # unit tests
make build                  # build binary as ./barghman
make releaser                # local goreleaser snapshot build
```

Run locally against a config file:

```bash
go run . -file example.toml
```

Note: `TestOneDayDifferentHours` in `bargheman_test.go` fails on a clean
checkout of `master` independent of most changes here (pre-existing,
unrelated to timezone/date-window logic) - don't assume you broke it unless
you touched that logic.

## Config files and secrets

- `config.toml`, `config_test.toml`, `prod_config.toml`, and any other
  `*.toml` files containing real SMTP passwords, domain credentials, or the
  `auth_token` are **local-only and must never be committed**. Only
  `example.toml` (with blank placeholder values) should be tracked in git.
- Never print or log full credentials when adding debug output; the app
  already logs the full parsed `Config` (including passwords) at debug
  level via `slog` - be careful when adding new debug logging of secrets to
  avoid making this worse, and don't paste real secrets into commit
  messages, PRs, or this repo's tracked files.

## Conventions

- Logging uses `log/slog` with structured key/value pairs, e.g.
  `slog.Error("message", "error", err, "field", value)`. Follow this style
  instead of `fmt.Errorf`-only error reporting when logging failures.
- Network transports (`sendSMTP`, `sendEWS`) must close every connection
  they open on every code path, including error returns (`defer
  conn.Close()` / `defer client.Close()`), to avoid leaking sockets against
  the mail server's per-source connection limits.
- New SMTP-like config fields go on the `SMTP` struct in `config.go` with a
  `toml:"..."` tag, and any new required combinations should be validated in
  `ParseConfig`.
- Keep comments purpose-explaining, not narration of what the next line
  literally does.

## Releasing

Releases are built by GoReleaser via `.github/workflows/release.yml`,
triggered by pushing a tag matching `v*.*.*`. The changelog GoReleaser
generates from git log can be supplemented with a manually written release
description when more context is useful (e.g. new features).
