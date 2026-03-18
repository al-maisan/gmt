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
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/al-maisan/gmt/config"
)

var rePlaceholder = regexp.MustCompile(`%[A-Z][A-Z0-9_]*%`)

// Message holds a fully prepared email ready for sending.
type Message struct {
	Name        string
	Address     string
	Body        string
	Subject     string
	Cc          []string
	Attachments []string
}

func substVars(recipient config.Recipient, text string) string {
	pairs := []string{
		"%EA%", recipient.Email,
		"%FN%", recipient.First,
		"%LN%", recipient.Last,
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
		recipient.Data = maps.Clone(recipient.Data)

		cc := resolveOverride(cfg.Cc, recipient.Data, "CC")
		attachments := resolveOverride(cfg.Attachments, recipient.Data, "AS")
		delete(recipient.Data, "CC")
		delete(recipient.Data, "AS")

		subject := substVars(recipient, cfg.Subject)
		body := substVars(recipient, template)

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

func resolveOverride(global []string, data map[string]string, key string) []string {
	val, ok := data[key]
	if !ok {
		return slices.Clone(global)
	}
	if strings.HasPrefix(val, "+") {
		return append(slices.Clone(global), splitTrim(val[1:])...)
	}
	return splitTrim(val)
}

func splitTrim(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
