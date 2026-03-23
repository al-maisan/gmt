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

// rePlaceholder matches template variables in the form %KEY% where KEY is
// one or more uppercase letters, digits, or underscores.
var rePlaceholder = regexp.MustCompile(`%[A-Z][A-Z0-9_]*%`)

const (
	placeholderEmail     = "%EA%"
	placeholderFirstName = "%FN%"
	placeholderLastName  = "%LN%"

	overrideKeyCc     = "CC"
	overrideKeyAttach = "AS"
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
		// Clone to avoid mutating the original config during override processing.
		recipient.Data = maps.Clone(recipient.Data)

		cc := resolveOverride(cfg.Cc, recipient.Data, overrideKeyCc)
		attachments := resolveOverride(cfg.Attachments, recipient.Data, overrideKeyAttach)
		delete(recipient.Data, overrideKeyCc)
		delete(recipient.Data, overrideKeyAttach)

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
// If data contains key, its value replaces the global list; a leading "+"
// means append to the global list instead.
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

// splitTrim splits s on commas and trims whitespace from each element,
// discarding empty entries.
func splitTrim(s string) []string {
	var result []string
	for item := range strings.SplitSeq(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
