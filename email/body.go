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

package email

import (
	"fmt"
	"strings"

	"github.com/al-maisan/gmt/config"
)

type Mail struct {
	Recipient string
	Body      string
	Cmdline   []string
}

func substVars(recipient config.Recipient, text string) (result string) {
	result = strings.Replace(text, "%EA%", recipient.Email, -1)
	result = strings.Replace(result, "%FN%", recipient.First, -1)
	result = strings.Replace(result, "%LN%", recipient.Last, -1)
	for k, v := range recipient.Data {
		result = strings.Replace(result, fmt.Sprintf("%%%s%%", k), v, -1)
	}
	return
}

func PrepMails(cfg config.Data, template string) (mails []Mail) {
	mails = make([]Mail, 0, len(cfg.Recipients))
	for _, recipient := range cfg.Recipients {
		subject := substVars(recipient, cfg.Subject)
		mail := Mail{
			Cmdline:   prepMUAArgs(cfg, recipient.Data, subject, recipient.Email),
			Recipient: recipient.Email,
			Body:      prepBody(cfg, recipient, subject, substVars(recipient, template)),
		}
		mails = append(mails, mail)
	}
	return
}

func prepBody(cfg config.Data, recipient config.Recipient, subject string, body string) string {
	if cfg.MailProg != "sendmail" {
		return body
	}
	lines := []string{fmt.Sprintf("To: %s", recipient.Email)}
	if subject != "" {
		lines = append(lines, fmt.Sprintf("Subject: %s", subject))
	}
	if cfg.Cc != nil {
		lines = append(lines, fmt.Sprintf("Cc: %s", strings.Join(cfg.Cc, ", ")))
	}
	if cfg.From != "" {
		lines = append(lines, fmt.Sprintf("From: %s", cfg.From))
	}
	if cfg.ReplyTo != "" {
		lines = append(lines, fmt.Sprintf("Reply-To: %s", cfg.ReplyTo))
	}
	lines = append(lines, fmt.Sprintf("X-Mailer: gmt, version %s, https://301.mx/gmt", cfg.Version))

	header := strings.Join(lines, "\n")
	return strings.Join([]string{header, body}, "\n\n")
}
