// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019-2025  "Muharem Hrnjadovic" <muharem@linux.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package email

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/al-maisan/gmt/config"
	mail "github.com/wneessen/go-mail"
)

// logf writes a formatted message to w, ignoring write errors
// (output is best-effort and must not interrupt the send loop).
func logf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format, args...) //nolint:errcheck
}

// SendOptions controls rate limiting and retry behavior.
type SendOptions struct {
	Delay      time.Duration // delay between messages
	Retries    int           // max retry attempts per message
	RetryDelay time.Duration // backoff between retries
}

// Sender abstracts the ability to send an email message.
type Sender interface {
	Send(msg *mail.Msg) error
	Close() error
}

// SMTPCredentials holds the SMTP server connection parameters.
type SMTPCredentials struct {
	Host     string
	Port     int
	User     string
	Password string
}

// LoadSMTPCredentials reads SMTP credentials from environment variables.
// Returns an error listing any missing variables.
func LoadSMTPCredentials() (SMTPCredentials, error) {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SENDER_EMAIL")
	password := os.Getenv("SENDER_PASSWORD")

	var missing []string
	if host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if portStr == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if user == "" {
		missing = append(missing, "SENDER_EMAIL")
	}
	if password == "" {
		missing = append(missing, "SENDER_PASSWORD")
	}
	if len(missing) > 0 {
		return SMTPCredentials{}, fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return SMTPCredentials{}, fmt.Errorf("SMTP_PORT must be a valid integer, got %q", portStr)
	}

	return SMTPCredentials{Host: host, Port: port, User: user, Password: password}, nil
}

// NewSMTPSender creates a connected SMTP sender using the given credentials.
// The caller must call Close when done.
func NewSMTPSender(creds SMTPCredentials) (Sender, error) {
	client, err := mail.NewClient(creds.Host,
		mail.WithPort(creds.Port),
		mail.WithUsername(creds.User),
		mail.WithPassword(creds.Password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client for %s:%d: %w", creds.Host, creds.Port, err)
	}

	if err := client.DialWithContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to %s:%d: %w", creds.Host, creds.Port, err)
	}

	return &smtpSender{client: client}, nil
}

// smtpSender wraps a go-mail Client to implement Sender.
type smtpSender struct {
	client *mail.Client
}

func (s *smtpSender) Send(msg *mail.Msg) error { return s.client.Send(msg) }
func (s *smtpSender) Close() error             { return s.client.Close() }

// SendResult holds the outcome of a bulk send operation.
type SendResult struct {
	Sent   int
	Failed int
}

// SendAll sends all prepared messages using the given sender.
// Progress is written to w; per-message errors do not stop the batch.
// Behavior is controlled by opts: delay between messages, retry count, and
// retry backoff duration.
func SendAll(w io.Writer, sender Sender, cfg config.MailConfig, msgs []Message, opts SendOptions) SendResult {
	total := len(msgs)
	width := len(fmt.Sprintf("%d", total))
	var result SendResult
	for i, m := range msgs {
		prefix := fmt.Sprintf("[%*d/%*d]", width, i+1, width, total)

		if err := sendOne(w, sender, cfg, m, prefix, opts); err != nil {
			result.Failed++
		} else {
			result.Sent++
		}

		if opts.Delay > 0 && i < total-1 {
			time.Sleep(opts.Delay)
		}
	}
	return result
}

// sendOne prepares and sends a single message, with retries.
// Returns nil on success, or the final error on failure.
func sendOne(w io.Writer, sender Sender, cfg config.MailConfig, m Message, prefix string, opts SendOptions) error {
	recipient := fmt.Sprintf("%s <%s>", m.Name, m.Address)

	msg, err := createMessage(cfg.From, m.Name, m.Address, m.Cc, cfg.ReplyTo, m.Subject, m.Body)
	if err != nil {
		logf(w,"%s ! %s (failed to create: %v)\n", prefix, recipient, err)
		return err
	}

	if err := attachFiles(msg, m.Attachments); err != nil {
		logf(w,"%s ! %s (failed to attach: %v)\n", prefix, recipient, err)
		return err
	}

	if err := sendWithRetry(w, sender, msg, prefix, recipient, opts); err != nil {
		return err
	}

	logf(w,"%s - %s\n", prefix, recipient)
	if len(m.Cc) > 0 {
		logf(w,"  Cc: %s\n", strings.Join(m.Cc, ", "))
	}
	if len(m.Attachments) > 0 {
		logf(w,"  Attachments: %s\n", strings.Join(m.Attachments, ", "))
	}
	return nil
}

// sendWithRetry attempts to send msg, retrying up to opts.Retries times
// with opts.RetryDelay between attempts.
func sendWithRetry(w io.Writer, sender Sender, msg *mail.Msg, prefix, recipient string, opts SendOptions) error {
	var err error
	for attempt := range opts.Retries + 1 {
		if attempt > 0 {
			logf(w,"%s   retrying %s...\n", prefix, recipient)
			if opts.RetryDelay > 0 {
				time.Sleep(opts.RetryDelay)
			}
		}
		err = sender.Send(msg)
		if err == nil {
			return nil
		}
	}
	logf(w,"%s ! %s (failed to send: %v)\n", prefix, recipient, err)
	return err
}

// createMessage builds a single email message with the given headers and body.
func createMessage(from, toName, toAddr string, cc []string, replyTo, subject, body string) (*mail.Msg, error) {
	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return nil, fmt.Errorf("invalid From address %q: %w", from, err)
	}
	if err := m.AddToFormat(toName, toAddr); err != nil {
		return nil, fmt.Errorf("invalid To address %q: %w", toAddr, err)
	}
	if len(cc) > 0 {
		if err := m.Cc(cc...); err != nil {
			return nil, fmt.Errorf("invalid Cc address(es) %v: %w", cc, err)
		}
	}
	if replyTo != "" {
		if err := m.ReplyTo(replyTo); err != nil {
			return nil, fmt.Errorf("invalid Reply-To address %q: %w", replyTo, err)
		}
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)
	return m, nil
}

// attachFiles adds the given file paths as attachments to msg, verifying
// each file exists before attaching.
func attachFiles(msg *mail.Msg, attachments []string) error {
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment %s: %w", path, err)
		}
		msg.AttachFile(path)
	}
	return nil
}
