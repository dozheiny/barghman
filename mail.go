package main

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"
)

var (
	MailHeadersFormat = "From: %s <%s>\r\n" + // Name and Email
		"To: %s\r\n" + // To.
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
		"ORGANIZER;CN=Barghman:mailto:%s\r\n" // Organizer.

	CalendarAttendanceFormat = "ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:%s\r\n"

	PlainTextBodyFormat = "--%s\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		"Dear colleague,\r\n\r\n" +
		"Please note the planned power outages below.\r\n" +
		"You can also add them directly to your calendar using the attached invitation.\r\n\r\n"

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

	default:
		return nil, fmt.Errorf("invalid auth method")
	}

	return &Mail{Auth: auth, Config: config, Loc: loc}, nil
}

func (m *Mail) Send(data []Data, recipients []string) error {
	boundary := generateBoundary()

	var content strings.Builder
	if _, err := content.WriteString(fmt.Sprintf(MailHeadersFormat,
		m.Config.Username,
		m.Config.Mail,
		strings.Join(recipients, ","),
		boundary,
	)); err != nil {
		slog.Error("Failed to write string", "error", err)
		return err
	}

	if _, err := content.WriteString(fmt.Sprintf(PlainTextBodyFormat, boundary)); err != nil {
		slog.Error("Failed to write plain text body", "error", err)
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
			fmt.Sprintf("%d@barghman.net", d.OutageNumber),
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

	return smtp.SendMail(fmt.Sprintf("%s:%s", m.Config.Address, m.Config.Port), m.Auth, m.Config.Mail, recipients, []byte(cont))
}
