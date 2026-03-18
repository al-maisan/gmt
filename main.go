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
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
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

	cfg := loadConfig(*configPath)
	mails := prepMails(&cfg, *templatePath)

	for _, w := range cfg.Warnings {
		log.Printf("Warning: %s", w)
	}

	if *doDryRun {
		printDryRun(mails)
		os.Exit(0)
	}

	creds, err := email.LoadSMTPCredentials()
	if err != nil {
		log.Fatalf("SMTP configuration error: %v", err)
	}

	fmt.Println("\nSending emails now..")
	result, err := email.SendAll(creds, cfg, mails)
	if err != nil {
		log.Fatalf("SMTP error: %v", err)
	}
	fmt.Printf("\nDone: %d sent, %d failed, %d total\n", result.Sent, result.Failed, result.Sent+result.Failed)

	if result.Failed > 0 {
		os.Exit(6)
	}
}

func loadConfig(path string) config.Data {
	bs, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file %q: %v", path, err)
	}

	c, err := config.New(bs)
	if err != nil {
		log.Fatalf("Failed to parse config file %q: %v", path, err)
	}

	cfg, err := c.Parse()
	if err != nil {
		log.Fatalf("Invalid config file %q: %v", path, err)
	}

	return cfg
}

func prepMails(cfg *config.Data, templatePath string) []email.Mail {
	bs, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Failed to read template file %q: %v", templatePath, err)
	}

	mails := email.PrepMails(cfg, string(bs))
	if len(mails) == 0 {
		log.Print("Warning: no recipients found in config file")
		os.Exit(0)
	}

	return mails
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
