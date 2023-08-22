// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019  "Muharem Hrnjadovic" <gmt@lbox.cc>
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
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
	"github.com/go-mail/mail"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func help() {
	fmt.Fprintf(flag.CommandLine.Output(), "\n%s, version %s\nThis tool sends emails in bulk based on a template and a config file\n\n", filepath.Base(os.Args[0]), version())
	flag.PrintDefaults()
}

func version() string { return "0.2.1" }

func main() {

	flag.Usage = help
	configPath := flag.String("config-path", "", "path to the config file")
	doDryRun := flag.Bool("dry-run", false, "show what would be done but execute no action")
	templatePath := flag.String("template-path", "", "path to the template file")
	doSampleConfig := flag.Bool("sample-config", false, "output sample configuration to stdout")
	doSampleTemplate := flag.Bool("sample-template", false, "output sample template to stdout")

	flag.Parse()

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
	bytes, err := ioutil.ReadFile(*configPath)
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
	bytes, err = ioutil.ReadFile(*templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read template file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	cfg.Version = version()
	// prepare the emails, substitute vars in subject & body
	mails := email.PrepMails(cfg, string(bytes))

	// is this a dry run? print what would be done if so and exit
	if *doDryRun == true {
		for _, mail := range mails {
			fmt.Printf("--\n%s\n%s\n%s\n", mail.Recipient, mail.Subject, mail.Body)
		}
		os.Exit(0)
	}

	send(cfg, mails)
}

func send(cfg config.Data, mails []email.Mail) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	from := os.Getenv("SENDER_EMAIL")
	password := os.Getenv("SENDER_PASSWORD")

	fmt.Println("\nSending emails now..")
	for _, mail := range mails {
		to, err := sendEmailWithAttachments(from, password, smtpHost, smtpPort, cfg, mail)
		if err == nil {
			fmt.Printf("- %s\n", to)
		} else {
			fmt.Printf("! %s (failed to send)\n", to)
		}
	}
	return
}

func sendEmailWithAttachments(
	from, password, smtpHost string, smtpPort int, cfg config.Data, email email.Mail) (string, error) {
	msg := createEmailMessage(cfg.From, email.Recipient, cfg.Cc, cfg.ReplyTo, email.Subject, email.Body)

	err := addAttachments(msg, cfg.Attachments)
	if err != nil {
		log.Errorf("failed to prep attachments for %s, %v", email.Recipient, err)
		return email.Recipient, err
	}

	recipients := append(append([]string{}, email.Recipient), cfg.Cc...)
	log.Printf("\n%s", recipients)

	d := mail.NewDialer(smtpHost, smtpPort, from, password)
	d.DialAndSend(msg)
	if err != nil {
		log.Errorf("failed to send email for %s, %v", email.Recipient, err)
		return email.Recipient, err
	}

	return email.Recipient, nil
}

func createEmailMessage(from, to string, cc []string, replyTo, subject, body string) *mail.Message {
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	if len(cc) > 0 {
		m.SetHeader("Cc", fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ",")))
	}
	if replyTo != "" {
		m.SetHeader("Reply-To", replyTo)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	return m
}

func createEmailMessage2(from, to string, cc []string, replyTo, subject, body string) *bytes.Buffer {
	msg := &bytes.Buffer{}

	headers := fmt.Sprintf("From: %s\r\n", from)
	headers += fmt.Sprintf("To: %s\r\n", to)

	if len(cc) > 0 {
		headers += fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ","))
	}

	headers += fmt.Sprintf("Reply-To: %s\r\n", replyTo)
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: multipart/mixed; boundary=boundary123\r\n"
	headers += "\r\n"

	msg.Write([]byte(headers))

	msg.Write([]byte("--boundary123\r\n"))
	msg.Write([]byte("Content-Type: text/plain; charset=utf-8\r\n"))
	msg.Write([]byte("\r\n"))
	msg.Write([]byte(body))
	msg.Write([]byte("\r\n"))

	return msg
}

func addAttachments(msg *mail.Message, attachments []string) error {
	for _, attachmentPath := range attachments {
		_, err := ioutil.ReadFile(attachmentPath)
		if err != nil {
			log.Errorf("failed to read attachment %s, %v", attachmentPath, err)
			return err
		}

		msg.Attach(attachmentPath)
	}

	return nil
}
