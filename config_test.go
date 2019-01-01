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

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadDefault(t *testing.T) {
	Convey("Load from good data sources", t, func() {
		cfg, err := NewConfig([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
#attachment-path=/tmp
#encrypt-attachments=true
sender-email=rts@example.com
sender-name=Frodo Baggins
#Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
[attachments]
jd@example.com=01.pdf
mm@gmail.com=02.pdf
`),
		)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)

		// Validate values make sure all sources are loaded correctly
		So(cfg.MailProg, ShouldEqual, "gnu-mail")
		So(cfg.AttachmentPath, ShouldEqual, "")
		So(cfg.EncryptAttachments, ShouldEqual, false)
		So(cfg.SenderEmail, ShouldEqual, "rts@example.com")
		So(cfg.SenderName, ShouldEqual, "Frodo Baggins")
		So(len(cfg.Cc), ShouldEqual, 0)
	})
}
func TestLoadFull(t *testing.T) {
	Convey("Load from good data sources", t, func() {
		cfg, err := NewConfig([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
attachment-path=/tmp
encrypt-attachments=true
sender-email=rts@example.com
sender-name=Frodo Baggins
Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
[attachments]
jd@example.com=01.pdf
mm@gmail.com=02.pdf
`),
		)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)

		// Validate values make sure all sources are loaded correctly
		So(cfg.MailProg, ShouldEqual, "gnu-mail")
		So(cfg.AttachmentPath, ShouldEqual, "/tmp")
		So(cfg.EncryptAttachments, ShouldEqual, true)
		So(cfg.SenderEmail, ShouldEqual, "rts@example.com")
		So(cfg.SenderName, ShouldEqual, "Frodo Baggins")
		So(len(cfg.Cc), ShouldEqual, 2)
		So(cfg.Cc, ShouldResemble, []string{"weirdo@nsb.gov", "cc@example.com"})
	})
}
