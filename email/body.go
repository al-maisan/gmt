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


func SubstVars(cfg config.Data, template string) (bodies map[string]string) {
	bodies = make(map[string]string)
	for _, recipient := range cfg.Recipients {
		body := strings.Replace(template, "%EA%", recipient.Email, -1)
		body = strings.Replace(body, "%FN%", recipient.First, -1)
		body = strings.Replace(body, "%LN%", recipient.Last, -1)
		for k, v := range recipient.Data {
			body = strings.Replace(body, fmt.Sprintf("%%%s%%", k), v, -1)
		}
		bodies[recipient.Email] = body
	}
	return
}
