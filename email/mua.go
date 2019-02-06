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
	"regexp"
	"strings"

	"github.com/al-maisan/gmt/config"
)

// `prepMUAArgs` is called for each email recipient. It converts global
// configuration data (`cfg`) and per-recipient configuration variables
// (`prdata`) to mail user agent (MUA) command line arguments.
func prepMUAArgs(cfg config.Data, prdata map[string]string, subject string, recipient string) (args []string) {
	if prccv, ok := prdata["Cc"]; ok {
		re := regexp.MustCompile("\\s*,\\s*")
		if strings.HasPrefix(prccv, "+") {
			val := strings.Trim(prccv[1:], " 	")
			cfg.Cc = append(cfg.Cc, re.Split(val, -1)...)
		} else {
			cfg.Cc = re.Split(prccv, -1)
		}
	}

	args = []string{cfg.MailProg}
	if cfg.MailProg == "sendmail" {
		args = append(args, "-t")
	} else if cfg.MailProg == "mailx" {
		if cfg.Cc != nil {
			args = append(args, "-c", strings.Join(cfg.Cc, ","))
		}
		if cfg.From != "" {
			args = append(args, "-S", fmt.Sprintf("from=%s", cfg.From))
		}
		if cfg.ReplyTo != "" {
			args = append(args, "-S", fmt.Sprintf("replyto=%s", cfg.ReplyTo))
		}
		if cfg.Attachments != nil {
			for _, path := range cfg.Attachments {
				args = append(args, "-a", path)
			}
		}
	} else {
		if cfg.Cc != nil {
			args = append(args, "-a", fmt.Sprintf("Cc: %s", strings.Join(cfg.Cc, ", ")))
		}
		if cfg.From != "" {
			args = append(args, "-a", fmt.Sprintf("From: %s", cfg.From))
		}
		if cfg.ReplyTo != "" {
			args = append(args, "-a", fmt.Sprintf("Reply-To: %s", cfg.ReplyTo))
		}
		if cfg.Attachments != nil {
			for _, path := range cfg.Attachments {
				args = append(args, "-A", path)
			}
		}
	}
	if cfg.MailProg != "sendmail" {
		args = append(args, "-s", subject, recipient)
	}
	return
}
