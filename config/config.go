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
	"strconv"

	"github.com/go-ini/ini"
)

type Recipient struct {
	Email string
	First string
	Last string
	Data map[string]string
}

type Attachment struct {
	Email string
	Path string
}

type General struct {
	MailProg string
	AttachmentPath string
	EncryptAttachments bool
	SenderEmail string
	SenderName string
	Cc []string
	Recipients []Recipient
	Attachments []Attachment
}

func New(bs []byte) (result General, err error) {
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

	// optional keys
	if val, ok := keys["attachment-path"]; ok {
		result.AttachmentPath = val
	}
	if val, ok := keys["encrypt-attachments"]; ok {
		result.EncryptAttachments, err = strconv.ParseBool(val)
		if err != nil {
			return
		}
	}
	if val, ok := keys["sender-email"]; ok {
		result.SenderEmail = val
	}
	if val, ok := keys["sender-name"]; ok {
		result.SenderName = val
	}
	if val, ok := keys["Cc"]; ok {
		re := regexp.MustCompile("\\s*,\\s*")
		result.Cc = re.Split(val, -1)
	}
	return
}

func ParseRecipients(cfg *ini.File) (recipients []Recipient, err error) {
	sec, err := cfg.GetSection("recipients")
	if err != nil {
		return
	}

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
			Last: names[1],
			Data: data,
		}
		recipients = append(recipients, recipient)
	}
	return
}
