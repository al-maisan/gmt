// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019-2023  "Muharem Hrnjadovic" <gmt@lbox.cc>
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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
	"github.com/go-mail/mail"
	"github.com/joho/godotenv"
)

func help() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "\n%s, version %s\nThis tool sends emails in bulk based on a template and a config file\n\n", filepath.Base(os.Args[0]), version())
	flag.PrintDefaults()
}

var (
	// Set via -ldflags at build time
	gitCommit = "unknown"
	buildDate = "unknown"
)

func version() string { return "0.2.1-" + gitCommit + " (" + buildDate + ")" }

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to load .env file: %v\n", err)
	}

	flag.Usage = help
	configPath := flag.String("config-path", "", "path to the config file")
	doDryRun := flag.Bool("dry-run", false, "show what would be done but execute no action")
	templatePath := flag.String("template-path", "", "path to the template file")
	doSampleConfig := flag.Bool("sample-config", false, "output sample configuration to stdout")
	doSampleTemplate := flag.Bool("sample-template", false, "output sample template to stdout")
	doVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *doVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	if *doSampleConfig {
		fmt.Println(config.SampleConfig(version()))
		os.Exit(0)
	}

	if *doSampleTemplate {
		fmt.Println(config.SampleTemplate(version()))
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "*** Error: please specify config file!")
		flag.Usage()
		os.Exit(1)
	}
	if *templatePath == "" {
		fmt.Fprintln(os.Stderr, "*** Error: please specify template file!")
		flag.Usage()
		os.Exit(2)
	}

	// read the config file
	bytes, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	// parse the config file
	var cfg config.Data
	cfg, err = config.New(bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	// read the template file
	bytes, err = os.ReadFile(*templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read template file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	// prepare the emails, substitute vars in subject & body
	mails := email.PrepMails(cfg, string(bytes))

	if len(mails) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: no recipients found in config file")
		os.Exit(0)
	}

	// is this a dry run? print what would be done if so and exit
	if *doDryRun {
		for _, m := range mails {
			fmt.Printf("--\n\"%s\" <%s>\n", m.Name, m.Address)
			if len(m.Cc) > 0 {
				fmt.Printf("Cc: %s\n", strings.Join(m.Cc, ", "))
			}
			fmt.Printf("Subject: %s\n", m.Subject)
			if len(m.Attachments) > 0 {
				fmt.Printf("Attachments: %s\n", strings.Join(m.Attachments, ", "))
			}
			fmt.Printf("%s\n", m.Body)
		}
		os.Exit(0)
	}

	failed := send(cfg, mails)
	if failed > 0 {
		os.Exit(6)
	}
}

func send(cfg config.Data, mails []email.Mail) int {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	from := os.Getenv("SENDER_EMAIL")
	password := os.Getenv("SENDER_PASSWORD")

	if smtpHost == "" || smtpPortStr == "" || from == "" || password == "" {
		fatal("SMTP_HOST, SMTP_PORT, SENDER_EMAIL, and SENDER_PASSWORD environment variables must all be set")
	}
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		fatal("SMTP_PORT must be a valid integer, got %q: %v", smtpPortStr, err)
	}

	d := mail.NewDialer(smtpHost, smtpPort, from, password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	sender, err := d.Dial()
	if err != nil {
		fatal("failed to connect to SMTP server: %v", err)
	}
	defer func() {
		if err := sender.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close SMTP connection: %v\n", err)
		}
	}()

	fmt.Println("\nSending emails now..")
	var sent, failed int
	for _, m := range mails {
		recipient := fmt.Sprintf("%s <%s>", m.Name, m.Address)

		msg := createEmailMessage(cfg.From, m.Name, m.Address, m.Cc, cfg.ReplyTo, m.Subject, m.Body)

		if err := addAttachments(msg, m.Attachments); err != nil {
			fmt.Printf("! %s (failed to attach: %v)\n", recipient, err)
			failed++
			continue
		}

		if err := mail.Send(sender, msg); err != nil {
			fmt.Printf("! %s (failed to send: %v)\n", recipient, err)
			failed++
			continue
		}

		fmt.Printf("- %s\n", recipient)
		sent++
	}
	fmt.Printf("\nDone: %d sent, %d failed, %d total\n", sent, failed, sent+failed)
	return failed
}

func createEmailMessage(from, toName, toAddr string, cc []string, replyTo, subject, body string) *mail.Message {
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
	for _, attachmentPath := range attachments {
		if _, err := os.Stat(attachmentPath); err != nil {
			return fmt.Errorf("attachment %s: %w", attachmentPath, err)
		}
		msg.Attach(attachmentPath)
	}
	return nil
}
