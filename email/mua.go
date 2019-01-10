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

func PrepMUAArgs(cfg config.Data) (args []string) {
	args = make([]string, 0)
	if cfg.MailProg == "mailx" {
		if cfg.Cc != nil {
			args = append(args, []string{"-c", fmt.Sprintf("'%s'", strings.Join(cfg.Cc, ", "))}...)
		}
		if cfg.SenderEmail != "" {
			var sender string
			if cfg.SenderName != "" {
				sender = fmt.Sprintf("'%s <%s>'", cfg.SenderName, cfg.SenderEmail)
			} else {
				sender = cfg.SenderEmail
			}
			from := fmt.Sprintf("from=%s", sender)
			args = append(args, []string{"-S", from}...)
		}
		if cfg.ReplyTo != "" {
			replyto := fmt.Sprintf("replyto='%s'", cfg.ReplyTo)
			args = append(args, []string{"-S", replyto}...)
		}
	} else {
		if cfg.Cc != nil {
			args = append(args, []string{"-a", fmt.Sprintf("'Cc: %s'", strings.Join(cfg.Cc, ", "))}...)
		}
		if cfg.SenderEmail != "" {
			var sender string
			if cfg.SenderName != "" {
				sender = fmt.Sprintf("'From: %s <%s>'", cfg.SenderName, cfg.SenderEmail)
			} else {
				sender = fmt.Sprintf("'From: %s'", cfg.SenderEmail)
			}
			args = append(args, []string{"-a", sender}...)
		}
		if cfg.ReplyTo != "" {
			replyto := fmt.Sprintf("'Reply-To: %s'", cfg.ReplyTo)
			args = append(args, []string{"-a", replyto}...)
		}
	}
	return
}
