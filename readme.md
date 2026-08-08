# Barghman

Barghman is a service that connects to the Iran Power electricity provider and sends calendar emails in ICS format with your blackout schedules. It can run as a standalone command or as a scheduled service using cron jobs.

*Hope for the days when we don't need this fucking service for power outages.*

## Installation

### One-line installer (Linux / macOS / Windows MSYS2)

```bash
curl -fsSL https://github.com/dozheiny/barghman/releases/latest/download/install.sh | bash
```

Or with `wget`:

```bash
wget -qO- https://github.com/dozheiny/barghman/releases/latest/download/install.sh | bash
```

The script auto-detects your OS and architecture (`linux/darwin/windows`, `amd64/arm64`), downloads the correct release archive, and installs the binary to `/usr/local/bin` (or `%USERPROFILE%\bin` on Windows).

### AUR (Arch Linux)

```bash
yay -S barghman-git
```

Or manually:

```bash
git clone https://aur.archlinux.org/barghman-git.git
cd barghman-git && makepkg -si
```

### From source

```bash
git clone https://github.com/dozheiny/barghman.git
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

## Docker

Build the image:

```bash
make docker
# or: docker build -t barghman:latest .
```

Run with a mounted config (and optional cache volume so sent-outage state survives restarts):

```bash
docker run --rm \
  -v /path/to/config.toml:/etc/barghman/config.toml:ro \
  -v barghman-cache:/var/cache/barghman \
  barghman:latest
```

The image entrypoint is `barghman -file /etc/barghman/config.toml`. Use a non-empty `cron_job` in the config so the container keeps running on a schedule.

## Config File Format

### General Options

| Option      | Default | Description                                                                 |
| ----------- | ------- | --------------------------------------------------------------------------- |
| `log_level` | `0`     | Logger verbosity level.                                                      |
| `auth_token` | Authentication token provided by https://uiapi.saapa.ir |
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
| `transport`   | `smtp` (default) sends directly over SMTP. `ews` sends the same message through Exchange Web Services over HTTPS instead - useful when SMTP is blocked/throttled but EWS/OWA is reachable. |
| `ews_url`     | Only used when `transport = "ews"`. Overrides the default EWS endpoint (`https://<host>/EWS/Exchange.asmx`). |

When `transport = "ews"`, `username`/`password` are used for NTLM authentication against EWS (e.g. `DOMAIN\user`), and `auth_method` is ignored.

#### Proxy (optional)

To route SMTP (or EWS) traffic through a SOCKS5 proxy, add a `[smtp.<provider>.proxy]` sub-table:

| Option     | Description                                          |
| ---------- | ---------------------------------------------------- |
| `host`     | SOCKS5 proxy host.                                   |
| `port`     | SOCKS5 proxy port.                                   |
| `username` | Proxy username (leave empty if auth not required).   |
| `password` | Proxy password (leave empty if auth not required).   |

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

[smtp.gmail.proxy]
host = "127.0.0.1"
port = "1080"
username = ""   # omit if proxy needs no auth
password = ""
```

### Client Configuration

Each client represents a connection to an electricity service account.

| Option       | Description                                               |
| ------------ | --------------------------------------------------------- |
| `smtp` | Name of the `[smtp.<name>]` table to use for this client (e.g. `gmail` for `[smtp.gmail]`). |
| `bill_id`    | Unique identifier for your electricity bill.               |
| `bill_ids` | Unique identifiers for your electricity bills, This option added to avoid breaking changes here.|
| `recipients` | List of email addresses to send the calendar emails to. When running on a cron schedule, the config file is reloaded each cycle: newly added recipients get invites for already-known outages (only them), and same-day outage time changes trigger an update email to all current recipients. |

