package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

var (
	MailHeadersFormat = "From: %s <%s>\r\n" + // Name and Email
		"To: %s\r\n" + // To.
		"Bcc: %s\r\n" + // Bcc.
		"Subject: Power Outage\r\n" + // Subject.
		"MIME-Version: 1.0\r\n" + // MIME-Version.
		"Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n" // Boundary.

	CalendarHeaderContent = "--%s\r\n" +
		"Content-Type: text/calendar; method=REQUEST; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		"BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Blu//Barghman Calendar//EN\r\nCALSCALE:GREGORIAN\r\nMETHOD:REQUEST\r\n"

	CalendarFooterContent = "STATUS:CONFIRMED\r\nTRANSP:OPAQUE\r\nPRIORITY:5\r\nEND:VEVENT\r\n\r\n"

	CalendarEndContent = "END:VCALENDAR\r\n--%s--\r\n"

	CalendarBodyFormat = "BEGIN:VEVENT\r\n" +
		"UID:%s\r\n" + // Unique ID.
		"DTSTAMP:%s\r\n" + // When Event created.
		"DTSTART:%s\r\n" + // Start time.
		"DTEND:%s\r\n" + // End time.
		"SUMMARY:%s\r\n" + // Summary.
		"DESCRIPTION:%s\r\n" + // Event details.
		"LOCATION:%s\r\n" + // Location.
		"ORGANIZER;CN=Iliya:mailto:%s\r\n" // Organizer.

	CalendarAttendanceFormat = "ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:%s\r\n"

	timeFormat = "20060102T150405Z"
)

type Mail struct {
	Auth   smtp.Auth
	Config SMTP
	Loc    *time.Location
}

func NewMailClient(config SMTP, loc *time.Location) (*Mail, error) {
	var auth smtp.Auth
	switch config.AuthMethod {
	case smtpAuthMethodMD5:
		auth = smtp.CRAMMD5Auth(config.Username, config.Password)

	case smtpAuthMethodPlain:
		auth = smtp.PlainAuth(config.Identity, config.Username, config.Password, config.Address)

	case smtpAuthMethodCustom:
		auth = LoginAuth(config.Username, config.Password)

	default:
		return nil, fmt.Errorf("invalid auth method")
	}

	return &Mail{Auth: auth, Config: config, Loc: loc}, nil
}

func (m *Mail) Do(data []Data, recipients []string) error {
	boundary := generateBoundary()

	var content strings.Builder
	if _, err := content.WriteString(fmt.Sprintf(MailHeadersFormat,
		m.Config.From,
		m.Config.Mail,
		m.Config.Mail,
		strings.Join(recipients, ","),
		boundary,
	)); err != nil {
		slog.Error("Failed to write string", "error", err)
		return err
	}

	if _, err := content.WriteString(fmt.Sprintf(CalendarHeaderContent, boundary)); err != nil {
		slog.Error("Failed to write calendar header content", "error", err)
		return err
	}

	for _, d := range data {
		startDate, endDate, err := d.ParseTime(m.Loc)
		if err != nil {
			slog.Error("Failed to parse time", "error", err)
			continue
		}

		if _, err := content.WriteString(fmt.Sprintf(CalendarBodyFormat,
			fmt.Sprintf("%d", d.OutageNumber),
			time.Now().UTC().Format(timeFormat),
			startDate.UTC().Format(timeFormat),
			endDate.UTC().Format(timeFormat),
			d.Summary(),
			d.Description(),
			d.Address,
			m.Config.Mail,
		)); err != nil {
			slog.Error("Failed to write event body", "error", err)
			return err
		}

		for _, recipient := range recipients {
			if _, err := content.WriteString(fmt.Sprintf(CalendarAttendanceFormat, recipient)); err != nil {
				slog.Error("Failed to write recipient", "error", err)
				return err
			}
		}

		if _, err := content.WriteString(CalendarFooterContent); err != nil {
			slog.Error("Failed to write event-footer", "error", err)
			return err
		}
	}

	if _, err := content.WriteString(fmt.Sprintf(CalendarEndContent, boundary)); err != nil {
		slog.Error("Failed to write calendar end content", "error", err)
	}

	cont := content.String()
	slog.Debug("content generated", "content", cont)

	// return smtp.SendMail(fmt.Sprintf("%s:%s", m.Config.Address, m.Config.Port), m.Auth, m.Config.Mail, recipients, []byte(cont))
	return m.Send(cont, recipients)
}

func (m Mail) Send(msg string, recipients []string) error {
	conn, err := net.Dial("tcp", net.JoinHostPort(m.Config.Address, m.Config.Port))
	if err != nil {
		slog.Error("can't dial the server", "error", err, "address", m.Config.Address)
		return err
	}

	client, err := smtp.NewClient(conn, m.Config.Address)
	if err != nil {
		slog.Error("smtp new client failed", "error", err, "address", m.Config.Address)
		return err
	}

	if err := client.StartTLS(&tls.Config{ServerName: m.Config.Address, InsecureSkipVerify: m.Config.SkipTLS}); err != nil {
		slog.Error("can't start TLS", "error", err)
		return err
	}

	if err := client.Auth(m.Auth); err != nil {
		slog.Error("client auth failed", "error", err)
		return err
	}

	if err := client.Mail(m.Config.Mail); err != nil {
		slog.Error("client mail failed", "error", err)
		return err
	}

	if err := client.Rcpt(m.Config.Mail); err != nil {
		slog.Error("client rcpt failed", "error", err)
		return err
	}

	writer, err := client.Data()
	if err != nil {
		slog.Error("client data writer failed", "error", err)
		return err
	}

	defer writer.Close()

	if _, err := writer.Write([]byte(msg)); err != nil {
		slog.Error("writer.Write failed", "error", err)
		return err
	}

	return client.Quit()
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unkown fromServer")
		}
	}
	return nil, nil
}
