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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
	"github.com/go-mail/mail"
	"github.com/joho/godotenv"
)

var (
	// Set via -ldflags at build time
	gitCommit = "unknown"
	buildDate = "unknown"
)

func version() string { return "0.2.1-" + gitCommit + " (" + buildDate + ")" }

func help() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "\n%s, version %s\nThis tool sends emails in bulk based on a template and a config file\n\n", filepath.Base(os.Args[0]), version())
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to load .env file: %v", err)
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
		fmt.Println(config.SampleTemplate())
		os.Exit(0)
	}

	if *configPath == "" {
		log.Print("Error: -config-path flag is required")
		flag.Usage()
		os.Exit(1)
	}
	if *templatePath == "" {
		log.Print("Error: -template-path flag is required")
		flag.Usage()
		os.Exit(2)
	}

	cfg, mails := loadAndPrepare(*configPath, *templatePath)

	if *doDryRun {
		printDryRun(mails)
		os.Exit(0)
	}

	failed := send(cfg, mails)
	if failed > 0 {
		os.Exit(6)
	}
}

func loadAndPrepare(configPath, templatePath string) (config.Data, []email.Mail) {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file %q: %v", configPath, err)
	}

	c, err := config.New(cfgBytes)
	if err != nil {
		log.Fatalf("Failed to parse config file %q: %v", configPath, err)
	}

	cfg, err := c.ParseGeneral()
	if err != nil {
		log.Fatalf("Invalid [general] section in %q: %v", configPath, err)
	}

	cfg.Recipients, err = c.ParseRecipients()
	if err != nil {
		log.Fatalf("Invalid [recipients] section in %q: %v", configPath, err)
	}

	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Failed to read template file %q: %v", templatePath, err)
	}

	mails := email.PrepMails(cfg, string(tmplBytes))
	if len(mails) == 0 {
		log.Print("Warning: no recipients found in config file")
		os.Exit(0)
	}

	return cfg, mails
}

func printDryRun(mails []email.Mail) {
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
}

type smtpConfig struct {
	host     string
	port     int
	user     string
	password string
}

func loadSMTPConfig() smtpConfig {
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
		log.Fatalf("Missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("SMTP_PORT must be a valid integer, got %q", portStr)
	}

	return smtpConfig{host: host, port: port, user: user, password: password}
}

func send(cfg config.Data, mails []email.Mail) int {
	smtp := loadSMTPConfig()

	d := mail.NewDialer(smtp.host, smtp.port, smtp.user, smtp.password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	sender, err := d.Dial()
	if err != nil {
		log.Fatalf("Failed to connect to SMTP server %s:%d: %v", smtp.host, smtp.port, err)
	}
	defer func() {
		if err := sender.Close(); err != nil {
			log.Printf("Warning: failed to close SMTP connection to %s:%d: %v", smtp.host, smtp.port, err)
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
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment %s: %w", path, err)
		}
		msg.Attach(path)
	}
	return nil
}
