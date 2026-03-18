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

// Package email handles template substitution and per-recipient mail preparation.
package email

import (
	"fmt"
	"strings"

	"github.com/al-maisan/gmt/config"
)

// Mail holds a fully prepared email ready for sending.
type Mail struct {
	Name        string
	Address     string
	Body        string
	Subject     string
	Cc          []string
	Attachments []string
}

func substVars(recipient config.Recipient, text string) (result string) {
	result = strings.ReplaceAll(text, "%EA%", recipient.Email)
	result = strings.ReplaceAll(result, "%FN%", recipient.First)
	result = strings.ReplaceAll(result, "%LN%", recipient.Last)
	for k, v := range recipient.Data {
		result = strings.ReplaceAll(result, fmt.Sprintf("%%%s%%", k), v)
	}
	return
}

// PrepMails generates a Mail for each recipient by substituting template variables
// and resolving per-recipient Cc and attachment overrides.
func PrepMails(cfg config.Data, template string) (mails []Mail) {
	mails = make([]Mail, 0, len(cfg.Recipients))
	for _, recipient := range cfg.Recipients {
		// copy the Data map so we don't mutate the original
		data := make(map[string]string, len(recipient.Data))
		for k, v := range recipient.Data {
			data[k] = v
		}
		recipient.Data = data

		cc := resolveOverride(cfg.Cc, recipient.Data, "CC")
		attachments := resolveOverride(cfg.Attachments, recipient.Data, "AS")
		// remove Cc/As from Data so they are not used as template variables
		delete(recipient.Data, "CC")
		delete(recipient.Data, "AS")

		name := strings.TrimSpace(recipient.First + " " + recipient.Last)
		mail := Mail{
			Name:        name,
			Address:     recipient.Email,
			Subject:     substVars(recipient, cfg.Subject),
			Body:        substVars(recipient, template),
			Cc:          cc,
			Attachments: attachments,
		}
		mails = append(mails, mail)
	}
	return
}

// resolveOverride applies per-recipient override logic for Cc and Attachments.
// If the recipient has a value starting with "+", it is appended to the global list.
// Otherwise the recipient value replaces the global list entirely.
func resolveOverride(global []string, data map[string]string, key string) []string {
	val, ok := data[key]
	if !ok {
		if global == nil {
			return nil
		}
		result := make([]string, len(global))
		copy(result, global)
		return result
	}
	if strings.HasPrefix(val, "+") {
		result := make([]string, len(global))
		copy(result, global)
		return append(result, splitTrim(val[1:])...)
	}
	return splitTrim(val)
}

// splitTrim splits a comma-separated string and trims whitespace, skipping empty entries.
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
