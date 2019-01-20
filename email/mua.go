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

// `PrepMUAArgs` converts global configuration data to mail user agent (MUA)
// command line arguments.
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

// Return the index of the command line argument with the "Cc" header value or
// -1 if not present.
func findCcArgs(cmdline []string) (result int) {
	result = -1
	mailprog := cmdline[0]
	args := cmdline[1:]
	for idx, arg := range args {
		idx++
		if mailprog == "mailx" {
			if arg == "-c" {
				result = idx + 1
				break
			}
		} else {
			if arg == "-a" && strings.HasPrefix(args[idx], "'Cc: ") {
				result = idx + 1
				break
			}
		}
	}
	return
}

// `PostProcessMUAArgs` looks at per-recipient configuration data / variables
// and adds to the mail user agent (MUA) command line arguments if/as needed.
// In a first implementation we will support per-recipient additions XOR
// redefinitions of the 'Cc' header variable.
func PostProcessMUAArgs(data Data, cmdline []string) (result []string) {
	result = cmdline
	rcc, ok := data.RecipientVars["Cc"]
	if !ok {
		return
	}

	fmt.Println(rcc)
	return
}
