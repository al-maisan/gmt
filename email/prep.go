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

// Package email handles template substitution, per-recipient mail preparation,
// and SMTP delivery.
package email

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/al-maisan/gmt/config"
)

// rePlaceholder matches template variables in the form %KEY% where KEY is
// one or more uppercase letters, digits, or underscores.
var rePlaceholder = regexp.MustCompile(`%[A-Z][A-Z0-9_]*%`)

const (
	placeholderEmail     = "%EA%"
	placeholderFirstName = "%FN%"
	placeholderLastName  = "%LN%"
)

// Message holds a fully prepared email ready for sending.
type Message struct {
	Name        string
	Address     string
	Body        string
	Subject     string
	Cc          []string
	Attachments []string
}

// substituteVariables replaces placeholder tokens (%FN%, %LN%, %EA%, and
// any custom keys from recipient.Data) in text with their values.
func substituteVariables(recipient config.Recipient, text string) string {
	pairs := []string{
		placeholderEmail, recipient.Email,
		placeholderFirstName, recipient.First,
		placeholderLastName, recipient.Last,
	}
	for k, v := range recipient.Data {
		pairs = append(pairs, fmt.Sprintf("%%%s%%", k), v)
	}
	return strings.NewReplacer(pairs...).Replace(text)
}

// PrepMails generates a Message for each recipient by substituting template
// variables and resolving per-recipient Cc and attachment overrides.
// Warnings about unresolved placeholders are appended to cfg.Warnings.
func PrepMails(cfg *config.MailConfig, template string) []Message {
	mails := make([]Message, 0, len(cfg.Recipients))
	for _, recipient := range cfg.Recipients {
		cc := resolveOverride(cfg.Cc, recipient.Cc, recipient.CcExtra)
		attachments := resolveOverride(cfg.Attachments, recipient.Attachments, recipient.AttachmentsExtra)

		subject := substituteVariables(recipient, cfg.Subject)
		body := substituteVariables(recipient, template)

		if unresolved := rePlaceholder.FindAllString(subject, -1); len(unresolved) > 0 {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("recipient '%s': unresolved placeholder(s) in subject: %s", recipient.Email, strings.Join(unresolved, ", ")))
		}
		if unresolved := rePlaceholder.FindAllString(body, -1); len(unresolved) > 0 {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("recipient '%s': unresolved placeholder(s) in body: %s", recipient.Email, strings.Join(unresolved, ", ")))
		}

		name := strings.TrimSpace(recipient.First + " " + recipient.Last)
		mails = append(mails, Message{
			Name:        name,
			Address:     recipient.Email,
			Subject:     subject,
			Body:        body,
			Cc:          cc,
			Attachments: attachments,
		})
	}
	return mails
}

// resolveOverride returns the effective list for a field (Cc or attachments).
// If replace is set, it replaces global. If extra is set, it appends to global.
// Otherwise, global is returned as-is.
func resolveOverride(global, replace, extra []string) []string {
	if len(replace) > 0 {
		return replace
	}
	if len(extra) > 0 {
		return append(slices.Clone(global), extra...)
	}
	return slices.Clone(global)
}

