// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019-2023  "Muharem Hrnjadovic" <gmt@lbox.cc>
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
	"fmt"
	"os"
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
	MailProg    string
	From        string
	ReplyTo     string
	Cc          []string
	Subject     string
	Version     string
	Recipients  []Recipient
	Attachments []string
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
	if val, ok := keys["subject"]; ok {
		result.Subject = val
	} else {
		err = errors.New("'subject' not configured!")
		return
	}

	// optional keys
	if val, ok := keys["From"]; ok {
		result.From = val
	}
	if val, ok := keys["Reply-To"]; ok {
		result.ReplyTo = val
	}
	if val, ok := keys["Cc"]; ok {
		re := regexp.MustCompile(`\s*,\s*`)
		result.Cc = re.Split(val, -1)
	}
	if val, ok := keys["attachments"]; ok {
		if result.MailProg == "sendmail" {
			err = errors.New("Cannot use 'sendmail' with attachments!")
			return
		}
		re := regexp.MustCompile(`\s*,\s*`)
		result.Attachments = re.Split(val, -1)
		if path, err2 := checkAttachments(result.Attachments); err2 != nil {
			err = fmt.Errorf("Attachment '%s' does not exist!", path)
			return
		}
	}

	var recipients *ini.Section
	recipients, err = cfg.GetSection("recipients")
	if err == nil {
		result.Recipients = parseRecipients(recipients)
	}
	return
}

func checkAttachments(attachments []string) (path string, err error) {
	for _, path = range attachments {
		if _, err = os.Lstat(path); err != nil {
			return
		}
	}
	return
}

func parseRecipients(sec *ini.Section) (recipients []Recipient) {
	re_pipe := regexp.MustCompile(`\s*\|\s*`)
	re_space := regexp.MustCompile(`\s+`)
	re_rdata := regexp.MustCompile(`\s*\:-\s*`)
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

func SampleConfig(version string) string {
	fs := `# gmt version %s
#
# anything that follows a hash is a comment
# email address is to the left of the '=' sign, first word after is
# the first name, the rest is the surname
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'sendmail' on Fedora
From="Frodo Baggins" <rts@example.com>
#Cc=weirdo@nsb.gov, cc@example.com
#Reply-To="John Doe" <jd@mail.com>
subject=Hello %%FN%%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
# The 'Cc' setting below *redefines* the global 'Cc' value above
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD|Cc:-bl@kf.io,info@ex.org
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
# The 'Cc' setting below *adds* to the global 'Cc' value above
daisy@example.com=Daisy Lila|ORG:-NASA|TITLE:-Dr.|Cc:-+inc@gg.org
# The 'As' setting below *redefines* the global 'attachments' value above
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD|As:-file1.txt,file2.md
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
# The 'As' setting below *adds* to the global 'attachments' value above
daisy@example.com=Daisy Lila|ORG:-NASA|TITLE:-Dr.|As:-+file3.pdf`
	return fmt.Sprintf(fs, version)
}

func SampleTemplate(version string) string {
	fs := `FN / LN / EA = first name / last name / email address

Hello %%FN%% // %%LN%%, how are things going at %%ORG%%?
this is your email: %%EA%% :)


Sent with gmt version %s, see https://301.mx/gmt for details.`
	return fmt.Sprintf(fs, version)
}
