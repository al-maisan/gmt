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
	args = []string{cfg.MailProg}
	if cfg.MailProg == "mailx" {
		if cfg.Cc != nil {
			args = append(args, "-c", fmt.Sprintf("'%s'", strings.Join(cfg.Cc, ", ")))
		}
		if cfg.From != "" {
			args = append(args, "-S", fmt.Sprintf("from='%s'", cfg.From))
		}
		if cfg.ReplyTo != "" {
			args = append(args, "-S", fmt.Sprintf("replyto='%s'", cfg.ReplyTo))
		}
	} else {
		if cfg.Cc != nil {
			args = append(args, "-a", fmt.Sprintf("'Cc: %s'", strings.Join(cfg.Cc, ", ")))
		}
		if cfg.From != "" {
			args = append(args, "-a", fmt.Sprintf("'From: %s'", cfg.From))
		}
		if cfg.ReplyTo != "" {
			args = append(args, "-a", fmt.Sprintf("'Reply-To: %s'", cfg.ReplyTo))
		}
	}
	return
}
