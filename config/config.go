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

// Package config handles parsing of gmt INI configuration files.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	rePipe  = regexp.MustCompile(`\s*\|\s*`)
	reSpace = regexp.MustCompile(`\s+`)
	reRdata = regexp.MustCompile(`\s*:-\s*`)
	reComma = regexp.MustCompile(`\s*,\s*`)
)

// Recipient holds a parsed recipient entry from the config file.
type Recipient struct {
	Email string
	First string
	Last  string
	Data  map[string]string
}

// Config wraps a parsed INI file and provides methods to extract
// the [general] and [recipients] sections.
type Config struct {
	file *ini.File
}

// MailConfig holds the fully parsed configuration for a mailing run.
type MailConfig struct {
	From        string
	ReplyTo     string
	Cc          []string
	Subject     string
	Recipients  []Recipient
	Attachments []string
	Warnings    []string
}

// New loads an INI-format configuration from the given bytes.
func New(bs []byte) (*Config, error) {
	f, err := ini.InsensitiveLoad(bs)
	if err != nil {
		return nil, err
	}
	return &Config{file: f}, nil
}

// Parse is a convenience method that parses both [general] and [recipients]
// sections in one call.
func (c *Config) Parse() (MailConfig, error) {
	cfg, err := c.ParseGeneral()
	if err != nil {
		return MailConfig{}, err
	}
	if err := c.ParseRecipients(&cfg); err != nil {
		return MailConfig{}, err
	}
	return cfg, nil
}

// ParseGeneral extracts the [general] section fields.
func (c *Config) ParseGeneral() (MailConfig, error) {
	sec, err := c.file.GetSection("general")
	if err != nil {
		return MailConfig{}, errors.New("section not found")
	}
	keys := sec.KeysHash()

	var result MailConfig

	// mandatory keys (all keys are lowercase due to InsensitiveLoad)
	val, ok := keys["subject"]
	if !ok {
		return MailConfig{}, errors.New("missing required key 'subject'")
	}
	result.Subject = val

	val, ok = keys["from"]
	if !ok {
		return MailConfig{}, errors.New("missing required key 'from'")
	}
	result.From = val

	// optional keys
	if val, ok := keys["reply-to"]; ok {
		result.ReplyTo = val
	}
	if val, ok := keys["cc"]; ok {
		result.Cc = reComma.Split(val, -1)
	}
	if val, ok := keys["attachments"]; ok {
		result.Attachments = reComma.Split(val, -1)
		if err := checkAttachments(result.Attachments); err != nil {
			return MailConfig{}, err
		}
	}

	return result, nil
}

// ParseRecipients extracts the [recipients] section into cfg.Recipients,
// appending any warnings about malformed entries to cfg.Warnings.
func (c *Config) ParseRecipients(cfg *MailConfig) error {
	sec, err := c.file.GetSection("recipients")
	if err != nil {
		return err
	}
	cfg.Recipients, cfg.Warnings = parseRecipients(sec)
	return nil
}

func checkAttachments(attachments []string) error {
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment '%s' does not exist", path)
		}
	}
	return nil
}

func parseRecipients(sec *ini.Section) ([]Recipient, []string) {
	var recipients []Recipient
	var warnings []string
	for email, v := range sec.KeysHash() {
		rdata := rePipe.Split(v, -1)
		if len(rdata) < 1 || strings.TrimSpace(rdata[0]) == "" {
			warnings = append(warnings, fmt.Sprintf("recipient '%s': empty name field", email))
			continue
		}
		names := reSpace.Split(rdata[0], 2)
		first := names[0]
		last := ""
		if len(names) > 1 {
			last = names[1]
		}
		data := make(map[string]string)
		for _, rdatum := range rdata[1:] {
			parts := reRdata.Split(rdatum, 2)
			if len(parts) != 2 {
				warnings = append(warnings, fmt.Sprintf("recipient '%s': malformed data field '%s' (expected KEY:-VALUE)", email, rdatum))
				continue
			}
			data[strings.ToUpper(parts[0])] = parts[1]
		}
		recipients = append(recipients, Recipient{
			Email: email,
			First: first,
			Last:  last,
			Data:  data,
		})
	}
	return recipients, warnings
}

//go:embed samples/config.ini
var sampleConfigContent string

//go:embed samples/template.eml
var sampleTemplateContent string

// SampleConfig returns a commented example configuration file.
func SampleConfig(version string) string {
	return "# gmt version " + version + "\n" + sampleConfigContent
}

// SampleTemplate returns an example email template demonstrating placeholder usage.
func SampleTemplate() string {
	return strings.TrimRight(sampleTemplateContent, "\n")
}
