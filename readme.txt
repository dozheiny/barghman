=============================== Barghman =============================== 

Barghman is a service that connects to your electricity provider and se-
nds calendar (ICS format) emails with your billing or usage schedule.
It can run as a standalone command or as a scheduled service using cron
jobs.

================================ Usage ================================= 
	barghman -file <config file>

-f <config file>: Path to your TOML configuration file.

=============================== Building =============================== 

1. Install the Golang compiler from https://go.dev.
2. Run the build command:
	`make build`

This will compile the barghman binary for your system.

=============================== Running ================================
1. Create a TOML configuration file, e.g., example.toml.
2. Update the file with your credentials, SMTP details, and client info-
rmation.

3. Run barghman with the config file:
	barghman -f example.toml

========================== Config File Format ========================== 

General Options

| Option      | Default | Description                                                                                        |
| ----------- | ------- | -------------------------------------------------------------------------------------------------- |
| `log_level` | `0`     | Logger verbosity level. Higher numbers show more detailed logs; negative numbers reduce verbosity. |
| `cron_job`  | `""`    | Cron expression for scheduling the service (e.g., `@daily`, `0 30 2 * * *`).                       |
| `use_cron`  | `false` | Set to `true` if you want Barghman to run automatically according to `cron_job`.                   |


SMTP Configuration

Each mail provider can be configured under [smtp.<provider>].

| Option        | Description                                                             |
| ------------- | ----------------------------------------------------------------------- |
| `mail`        | The sender email address.                                               |
| `host`        | SMTP server host.                                                       |
| `port`        | SMTP server port.                                                       |
| `username`    | Username for SMTP authentication.                                       |
| `password`    | Password for SMTP authentication.                                       |
| `auth_method` | Authentication method (`plain`, `cram-md5`, `custom`).				  |
| `identity`    | Optional identity for authentication.                                   |
| `skip_tls`    | Set to `true` to skip TLS verification (not recommended in production). |


Example:

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


Client Configuration:
Each client represents a connection to an electricity service account.
| Option       | Description                                                 |
| ------------ | ----------------------------------------------------------- |
| `bill_id`    | Unique identifier of your electricity bill.                 |
| `auth_token` | Authentication token provided by https://ios.barghman.com.  |
| `recipients` | List of email addresses to send the calendar emails to.     |


