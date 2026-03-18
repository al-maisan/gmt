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

// Package smtp handles SMTP connection management, email message construction,
// and bulk sending.
package smtp

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
	"github.com/go-mail/mail"
)

// Credentials holds the SMTP server connection parameters.
type Credentials struct {
	Host     string
	Port     int
	User     string
	Password string
}

// LoadCredentials reads SMTP credentials from environment variables.
// Returns an error listing any missing variables.
func LoadCredentials() (Credentials, error) {
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
		return Credentials{}, fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Credentials{}, fmt.Errorf("SMTP_PORT must be a valid integer, got %q", portStr)
	}

	return Credentials{Host: host, Port: port, User: user, Password: password}, nil
}

// Result holds the outcome of a bulk send operation.
type Result struct {
	Sent   int
	Failed int
}

// SendAll sends all prepared mails over a single SMTP connection.
func SendAll(creds Credentials, cfg config.Data, mails []email.Mail) (Result, error) {
	d := mail.NewDialer(creds.Host, creds.Port, creds.User, creds.Password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	sender, err := d.Dial()
	if err != nil {
		return Result{}, fmt.Errorf("failed to connect to %s:%d: %w", creds.Host, creds.Port, err)
	}
	defer func() {
		if err := sender.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close SMTP connection to %s:%d: %v\n", creds.Host, creds.Port, err)
		}
	}()

	var result Result
	for _, m := range mails {
		recipient := fmt.Sprintf("%s <%s>", m.Name, m.Address)

		msg := createMessage(cfg.From, m.Name, m.Address, m.Cc, cfg.ReplyTo, m.Subject, m.Body)

		if err := addAttachments(msg, m.Attachments); err != nil {
			fmt.Printf("! %s (failed to attach: %v)\n", recipient, err)
			result.Failed++
			continue
		}

		if err := mail.Send(sender, msg); err != nil {
			fmt.Printf("! %s (failed to send: %v)\n", recipient, err)
			result.Failed++
			continue
		}

		fmt.Printf("- %s\n", recipient)
		result.Sent++
	}
	return result, nil
}

func createMessage(from, toName, toAddr string, cc []string, replyTo, subject, body string) *mail.Message {
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetAddressHeader("To", toAddr, toName)
	if len(cc) > 0 {
		m.SetHeader("Cc", strings.Join(cc, ","))
	}
	if replyTo != "" {
		m.SetHeader("Reply-To", replyTo)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	return m
}

func addAttachments(msg *mail.Message, attachments []string) error {
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment %s: %w", path, err)
		}
		msg.Attach(path)
	}
	return nil
}
