=============================== Barghman =============================== 

Barghman is a service that connects to the Power electricity provider 
and sends calendar emails in ICS format with your blackout schedules.
It can run as a standalone command or as a scheduled service using cron 
jobs.

============================= Installation ==============================

	git clone git@github.com:dozheiny/barghman.git
	cd barghman
	make install


================================ Usage ================================= 
	barghman -file <config file>

-file <config file>: Path to your TOML configuration file.

=============================== Building =============================== 

1. Install the Golang compiler from https://go.dev.
2. Run the build command:
	`make build`

This will compile the barghman binary for your system.

=============================== Running ================================
1. Create a TOML configuration file, e.g., example.toml.
2. Update the file with your credentials, SMTP details, and client information.

3. Run barghman with the config file:
	barghman -file example.toml

========================== Config File Format ========================== 

General Options

| Option      | Default | Description                                                                      |
| ----------- | ------- | ---------------------------------------------------------------------------------|
| `log_level` | `0`     | Logger verbosity level.                                                          |
| `cron_job`  | `""`    | Cron expression for scheduling the service (e.g., `@daily`, `0 30 2 * * *`).     |
| `use_cron`  | `false` | Set to `true` if you want Barghman to run automatically according to `cron_job`. |


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
| `skip_tls`    | Set to `true` to skip TLS verification (not recommended for production).|


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
| `bill_id`    | Unique identifier for your electricity bill.                |
| `auth_token` | Authentication token provided by https://ios.barghman.com.  |
| `recipients` | List of email addresses to send the calendar emails to.     |




=============================== TO-DO ===============================

- [x] Make integration with systemd
- [ ] Add some documentation (man pages)
- [x] `make install,uninstall,clean` commands
- [ ] Add dockerfile
- [ ] Add it to AUR and brew
- [ ] Update readme with markdown
- [ ] Maybe support for custom notifications (SMS maybe?)
- [ ] Add support for multiple bill IDs
- [ ] Add support for multiple origin emails
- [ ] Move auth token outside
- [ ] Make file creator customizable (TOML or JSON)
- [ ] Move main file to cmd directory
- [ ] Add delete cache functionality 
- [ ] Add content to the email about what this email is, why you receive it,
	 and how to add it to calendars, etc.
- [ ] Add install.bash script (not only makefile, no required installed golang)