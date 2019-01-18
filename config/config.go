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

package config

import (
	"errors"
	"regexp"

	"github.com/go-ini/ini"
)

type Recipient struct {
	Email string
	First string
	Last  string
	Data  map[string]string
}

type Data struct {
	MailProg   string
	From       string
	ReplyTo    string
	Cc         []string
	Subject    string
	Recipients []Recipient
}

func New(bs []byte) (result Data, err error) {
	var cfg *ini.File
	cfg, err = ini.Load(bs)
	if err != nil {
		return
	}
	sec, err := cfg.GetSection("general")
	if err != nil {
		return
	}
	keys := sec.KeysHash()

	// mandatory keys
	if val, ok := keys["mail-prog"]; ok {
		result.MailProg = val
	} else {
		err = errors.New("'mail-prog' not configured!")
		return
	}
	if val, ok := keys["subject"]; ok {
		result.Subject = val
	} else {
		err = errors.New("'subject' not configured!")
		return
	}

	// optional keys
	if val, ok := keys["sender-email"]; ok {
		result.SenderEmail = val
	}
	if val, ok := keys["sender-name"]; ok {
		result.SenderName = val
	}
	if val, ok := keys["Reply-To"]; ok {
		result.ReplyTo = val
	}
	if val, ok := keys["Cc"]; ok {
		re := regexp.MustCompile("\\s*,\\s*")
		result.Cc = re.Split(val, -1)
	}

	var recipients *ini.Section
	recipients, err = cfg.GetSection("recipients")
	if err == nil {
		result.Recipients = parseRecipients(recipients)
	}
	return
}

func parseRecipients(sec *ini.Section) (recipients []Recipient) {
	re_pipe := regexp.MustCompile("\\s*\\|\\s*")
	re_space := regexp.MustCompile("\\s+")
	re_rdata := regexp.MustCompile("\\s*\\:-\\s*")
	for k, v := range sec.KeysHash() {
		// jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
		rdata := re_pipe.Split(v, -1)
		if len(rdata) < 1 {
			continue
		}
		names := re_space.Split(rdata[0], 2)
		data := make(map[string]string)
		for _, rdatum := range rdata[1:] {
			split_rdatum := re_rdata.Split(rdatum, 2)
			data[split_rdatum[0]] = split_rdatum[1]
		}
		recipient := Recipient{
			Email: k,
			First: names[0],
			Last:  names[1],
			Data:  data,
		}
		recipients = append(recipients, recipient)
	}
	return
}

func SampleConfig() string {
	return `# anything that follows a hash is a comment
# email address is to the left of the '=' sign, first word after is
# the first name, the rest is the surname
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
From=Frodo Baggins <rts@example.com>
#Cc=weirdo@nsb.gov, cc@example.com
#Reply-To=John Doe <jd@mail.com>
subject=Hello %FN%!
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!`
}

func SampleTemplate() string {
	return `FN / LN / EA = first name / last name / email address

Hello %FN% // %LN%, how are things going at %ORG%?
this is your email * 2: %EA%%EA%.`
}
