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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-ini/ini"
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

// Data holds the fully parsed configuration for a mailing run.
type Data struct {
	From        string
	ReplyTo     string
	Cc          []string
	Subject     string
	Recipients  []Recipient
	Attachments []string
}

// New loads an INI-format configuration from the given bytes.
func New(bs []byte) (*Config, error) {
	f, err := ini.InsensitiveLoad(bs)
	if err != nil {
		return nil, err
	}
	return &Config{file: f}, nil
}

// ParseGeneral extracts the [general] section fields into a Data struct.
func (c *Config) ParseGeneral() (Data, error) {
	sec, err := c.file.GetSection("general")
	if err != nil {
		return Data{}, errors.New("section not found")
	}
	keys := sec.KeysHash()

	var result Data

	// mandatory keys (all keys are lowercase due to InsensitiveLoad)
	val, ok := keys["subject"]
	if !ok {
		return Data{}, errors.New("missing required key 'subject'")
	}
	result.Subject = val

	val, ok = keys["from"]
	if !ok {
		return Data{}, errors.New("missing required key 'from'")
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
			return Data{}, err
		}
	}

	return result, nil
}

// ParseRecipients extracts the [recipients] section into a slice of Recipient.
func (c *Config) ParseRecipients() ([]Recipient, error) {
	sec, err := c.file.GetSection("recipients")
	if err != nil {
		return nil, err
	}
	return parseRecipients(sec), nil
}

func checkAttachments(attachments []string) error {
	for _, path := range attachments {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment '%s' does not exist", path)
		}
	}
	return nil
}

func parseRecipients(sec *ini.Section) []Recipient {
	var recipients []Recipient
	for k, v := range sec.KeysHash() {
		// jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
		rdata := rePipe.Split(v, -1)
		if len(rdata) < 1 {
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
				continue
			}
			data[strings.ToUpper(parts[0])] = parts[1]
		}
		recipients = append(recipients, Recipient{
			Email: k,
			First: first,
			Last:  last,
			Data:  data,
		})
	}
	return recipients
}

// SampleConfig returns a commented example configuration file.
func SampleConfig(version string) string {
	fs := `# gmt version %s
#
# anything that follows a hash is a comment
# email address is to the left of the '=' sign, first word after is
# the first name, the rest is the surname
#
# SMTP configuration should be set via environment variables or .env file:
# SMTP_HOST=smtp.example.com
# SMTP_PORT=587
# SENDER_EMAIL=your-email@example.com
# SENDER_PASSWORD=your-password
[general]
From="Frodo Baggins" <rts@example.com>
#Cc=weirdo@nsb.gov, cc@example.com
#Reply-To="John Doe" <jd@mail.com>
subject=Hello %%FN%%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
# The 'Cc' setting below *redefines* the global 'Cc' value above
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD|Cc:-bl@kf.io,info@ex.org
# The 'Cc' setting below *adds* to the global 'Cc' value above
daisy@example.com=Daisy Lila|ORG:-NASA|TITLE:-Dr.|Cc:-+inc@gg.org
# The 'As' setting below *redefines* the global 'attachments' value above
ab@example.com=Alice Brown|ORG:-MIT|As:-file1.txt,file2.md
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
# The 'As' setting below *adds* to the global 'attachments' value above
ef@example.com=Eve Foster|ORG:-CERN|TITLE:-Prof.|As:-+file3.pdf`
	return fmt.Sprintf(fs, version)
}

// SampleTemplate returns an example email template demonstrating placeholder usage.
func SampleTemplate() string {
	return `Dear %FN% %LN%,

How are things going at %ORG%?

Best regards`
}
