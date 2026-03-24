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

// Package config handles parsing of gmt TOML configuration files.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// tomlConfig mirrors the TOML file structure for decoding.
type tomlConfig struct {
	General    tomlGeneral    `toml:"general"`
	Recipients []tomlRecipient `toml:"recipients"`
}

// tomlGeneral holds the [general] section fields.
type tomlGeneral struct {
	From        string   `toml:"from"`
	Subject     string   `toml:"subject"`
	ReplyTo     string   `toml:"reply_to"`
	Cc          []string `toml:"cc"`
	Attachments []string `toml:"attachments"`
}

// tomlRecipient holds a single [[recipients]] entry.
type tomlRecipient struct {
	Email            string            `toml:"email"`
	First            string            `toml:"first"`
	Last             string            `toml:"last"`
	Data             map[string]string `toml:"data"`
	Cc               []string          `toml:"cc"`
	CcExtra          []string          `toml:"cc_extra"`
	Attachments      []string          `toml:"attachments"`
	AttachmentsExtra []string          `toml:"attachments_extra"`
}

// Recipient holds a parsed recipient entry from the config file.
type Recipient struct {
	Email            string
	First            string
	Last             string
	Data             map[string]string
	Cc               []string // replaces global Cc
	CcExtra          []string // appends to global Cc
	Attachments      []string // replaces global attachments
	AttachmentsExtra []string // appends to global attachments
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

// Parse decodes TOML-formatted configuration bytes into a MailConfig.
func Parse(bs []byte) (MailConfig, error) {
	var tc tomlConfig
	if _, err := toml.Decode(string(bs), &tc); err != nil {
		return MailConfig{}, fmt.Errorf("TOML syntax error: %w", err)
	}

	if tc.General.From == "" {
		return MailConfig{}, errors.New("missing required key 'from' in [general]")
	}
	if tc.General.Subject == "" {
		return MailConfig{}, errors.New("missing required key 'subject' in [general]")
	}

	if err := checkAttachments(tc.General.Attachments); err != nil {
		return MailConfig{}, err
	}

	if len(tc.Recipients) == 0 {
		return MailConfig{}, errors.New("no [[recipients]] entries found")
	}

	cfg := MailConfig{
		From:        tc.General.From,
		Subject:     tc.General.Subject,
		ReplyTo:     tc.General.ReplyTo,
		Cc:          tc.General.Cc,
		Attachments: tc.General.Attachments,
	}

	recipients, warnings := convertRecipients(tc.Recipients)
	cfg.Recipients = recipients
	cfg.Warnings = warnings

	return cfg, nil
}

// convertRecipients transforms TOML recipient entries into Recipient structs.
func convertRecipients(entries []tomlRecipient) ([]Recipient, []string) {
	var recipients []Recipient
	var warnings []string

	for _, e := range entries {
		if e.Email == "" {
			warnings = append(warnings, "recipient entry missing 'email' field, skipped")
			continue
		}
		if e.First == "" {
			warnings = append(warnings, fmt.Sprintf("recipient '%s': missing 'first' field", e.Email))
			continue
		}

		data := make(map[string]string, len(e.Data))
		for k, v := range e.Data {
			data[strings.ToUpper(k)] = v
		}

		recipients = append(recipients, Recipient{
			Email:            e.Email,
			First:            e.First,
			Last:             e.Last,
			Data:             data,
			Cc:               e.Cc,
			CcExtra:          e.CcExtra,
			Attachments:      e.Attachments,
			AttachmentsExtra: e.AttachmentsExtra,
		})
	}
	return recipients, warnings
}

// checkAttachments verifies that every path in attachments exists and is accessible.
func checkAttachments(attachments []string) error {
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("attachment not found: %s", path)
			}
			return fmt.Errorf("attachment not accessible: %s: %w", path, err)
		}
	}
	return nil
}

//go:embed samples/config.toml
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
