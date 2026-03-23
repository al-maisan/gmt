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

// SendOptions controls rate limiting and retry behavior.
type SendOptions struct {
	Delay      time.Duration // delay between messages
	Retries    int           // max retry attempts per message
	RetryDelay time.Duration // backoff between retries
}

// SendResult holds the outcome of a bulk send operation.
type SendResult struct {
	Sent   int
	Failed int
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

// BatchSender holds the per-batch state for delivering a set of messages.
type BatchSender struct {
	w       io.Writer
	sender  Sender
	from    string
	replyTo string
	opts    SendOptions
}

// NewBatchSender creates a BatchSender for delivering a batch of messages.
func NewBatchSender(w io.Writer, sender Sender, cfg config.MailConfig, opts SendOptions) *BatchSender {
	return &BatchSender{w: w, sender: sender, from: cfg.From, replyTo: cfg.ReplyTo, opts: opts}
}

// SendAll delivers all messages, logging progress to w.
// Per-message errors do not stop the batch.
func (sc *BatchSender) SendAll(msgs []Message) SendResult {
	total := len(msgs)
	width := len(fmt.Sprintf("%d", total))
	var result SendResult
	for i, m := range msgs {
		prefix := fmt.Sprintf("[%*d/%*d]", width, i+1, width, total)

		if err := sc.sendOne(m, prefix); err != nil {
			result.Failed++
		} else {
			result.Sent++
		}

		if sc.opts.Delay > 0 && i < total-1 {
			time.Sleep(sc.opts.Delay)
		}
	}
	return result
}

// --- internal ---

// logf writes a formatted message to w, ignoring write errors
// (output is best-effort and must not interrupt the send loop).
func logf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format, args...) //nolint:errcheck
}

// smtpSender wraps a go-mail Client to implement Sender.
type smtpSender struct {
	client *mail.Client
}

func (s *smtpSender) Send(msg *mail.Msg) error { return s.client.Send(msg) }
func (s *smtpSender) Close() error             { return s.client.Close() }

// sendOne prepares and sends a single message, with retries.
func (sc *BatchSender) sendOne(m Message, prefix string) error {
	recipient := fmt.Sprintf("%s <%s>", m.Name, m.Address)

	msg, err := createMessage(sc.from, m.Name, m.Address, m.Cc, sc.replyTo, m.Subject, m.Body)
	if err != nil {
		logf(sc.w, "%s ! %s (failed to create: %v)\n", prefix, recipient, err)
		return err
	}

	if err := attachFiles(msg, m.Attachments); err != nil {
		logf(sc.w, "%s ! %s (failed to attach: %v)\n", prefix, recipient, err)
		return err
	}

	if err := sc.sendWithRetry(msg, prefix, recipient); err != nil {
		return err
	}

	logf(sc.w, "%s - %s\n", prefix, recipient)
	if len(m.Cc) > 0 {
		logf(sc.w, "  Cc: %s\n", strings.Join(m.Cc, ", "))
	}
	if len(m.Attachments) > 0 {
		logf(sc.w, "  Attachments: %s\n", strings.Join(m.Attachments, ", "))
	}
	return nil
}

// sendWithRetry attempts to send msg, retrying up to opts.Retries times
// with opts.RetryDelay between attempts.
func (sc *BatchSender) sendWithRetry(msg *mail.Msg, prefix, recipient string) error {
	var err error
	for attempt := range sc.opts.Retries + 1 {
		if attempt > 0 {
			logf(sc.w, "%s   retrying %s...\n", prefix, recipient)
			if sc.opts.RetryDelay > 0 {
				time.Sleep(sc.opts.RetryDelay)
			}
		}
		err = sc.sender.Send(msg)
		if err == nil {
			return nil
		}
	}
	logf(sc.w, "%s ! %s (failed to send: %v)\n", prefix, recipient, err)
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
