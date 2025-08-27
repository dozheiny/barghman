# Barghman

Barghman is a service that connects to the Iran Power electricity provider and sends calendar emails in ICS format with your blackout schedules. It can run as a standalone command or as a scheduled service using cron jobs.

*Hope for the days when we don't need this fucking service for power outages.*

## Installation

[AUR](https://aur.archlinux.org/packages/barghman-git)

```bash
git clone git@github.com:dozheiny/barghman.git
cd barghman
make install
```

## Usage

```bash
barghman -file <config file>
```

**Options:**
- `-file <config file>`: Path to your TOML configuration file


If you wish for running barghman as a systemd service:
```bash
	systemctl --user daemon-reload
	systemctl --user enable barghman.service
```

## Building

1. Install the Go compiler from https://go.dev
2. Run the build command:
   ```bash
   make build
   ```

This will compile the barghman binary for your system.

## Running

1. Create a TOML configuration file (e.g., `example.toml`)
2. Update the file with your credentials, SMTP details, and client information
3. Run barghman with the config file:
   ```bash
   barghman -file example.toml
   ```

## Config File Format

### General Options

| Option      | Default | Description                                                                 |
| ----------- | ------- | --------------------------------------------------------------------------- |
| `log_level` | `0`     | Logger verbosity level.                                                      |
| `cron_job`  | `""`    | Cron expression for scheduling the service (e.g., `@daily`, `0 30 2 * * *`). Keep in mind that if cron_job is empty, it will run as a one-time job; otherwise, it will run as a cron job.|
| `wait_time` | `0` | The wait time specifies how many seconds to wait for each client or bill ID. This is necessary because the Barghman API imposes limits on its planned blackout endpoint.|  

### SMTP Configuration

Each mail provider can be configured under `[smtp.<provider>]`.

| Option        | Description                                                             |
| ------------- | ----------------------------------------------------------------------- |
| `mail`        | The sender email address.                                                |
| `host`        | SMTP server host.                                                        |
| `port`        | SMTP server port.                                                        |
| `username`    | Username for SMTP authentication.                                        |
| `password`    | Password for SMTP authentication.                                        |
| `auth_method` | Authentication method (`plain`, `cram-md5`, `custom`).                   |
| `identity`    | Optional identity for authentication.                                    |
| `skip_tls`    | Set to `true` to skip TLS verification. |

**Example:**

```toml
[smtp.gmail]
mail = "your-email@gmail.com"
host = "smtp.gmail.com"
port = "587"
username = "your-email@gmail.com"
password = "your-app-password"
auth_method = "plain"
identity = ""
skip_tls = true
```

### Client Configuration

Each client represents a connection to an electricity service account.

| Option       | Description                                               |
| ------------ | --------------------------------------------------------- |
| `smtp` | smtp is used to identify each SMTP configuration, allowing you to map specific SMTP configs to your clients. For example if your smtp config starts with `[smtp.gmail]` then the value of smtp_name should be gmail.|
| `bill_id`    | Unique identifier for your electricity bill.               |
| `bill_ids` | Unique identifiers for your electricity bills, This option added to avoid breaking changes here.|
| `auth_token` | Authentication token provided by https://uiapi.saapa.ir |
| `recipients` | List of email addresses to send the calendar emails to.    |

## TO-DO

- [x] Make integration with systemd
- [x] Add some documentation (man pages)
- [x] `make install`, `uninstall`, `clean` commands
- [x] Add it to AUR
- [x] Update README with Markdown
- [x] Add support for multiple bill IDs
- [x] Add support for multiple origin emails
- [x] Add delete cache functionality
- [ ] Add update mail functionality
- [ ] Add Dockerfile
- [ ] Add content to the email about what this email is, why you receive it, and how to add it to calendars, etc.
- [ ] Add install.bash script (not only Makefile, no required installed Go)
