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


type Data struct {
	Subject string
	Body string
}


func SubstVars(recipient config.Recipient, text string) (result string) {
	result = strings.Replace(text, "%EA%", recipient.Email, -1)
	result = strings.Replace(result, "%FN%", recipient.First, -1)
	result = strings.Replace(result, "%LN%", recipient.Last, -1)
	for k, v := range recipient.Data {
		result = strings.Replace(result, fmt.Sprintf("%%%s%%", k), v, -1)
	}
	return
}

func PrepMails(cfg config.Data, template string) (mails map[string]Data) {
	mails = make(map[string]Data)
	for _, recipient := range cfg.Recipients {
		data := Data{}
		data.Body = SubstVars(recipient, template)
		data.Subject = SubstVars(recipient, cfg.Subject)
		mails[recipient.Email] = data
	}
	return
}
