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
	"sort"
	"testing"

	"github.com/go-ini/ini"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadDefault(t *testing.T) {
	Convey("Load sample config, test general parts", t, func() {
		cfg, err := New([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
sender-email=rts@example.com
sender-name=Frodo Baggins
#Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)

		So(cfg.MailProg, ShouldEqual, "gnu-mail")
		So(cfg.SenderEmail, ShouldEqual, "rts@example.com")
		So(cfg.SenderName, ShouldEqual, "Frodo Baggins")
		So(len(cfg.Cc), ShouldEqual, 0)
		expected := []Recipient {
			Recipient{
				Email: "jd@example.com",
				First: "John",
				Last: "Doe Jr.",
				Data: map[string]string {
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			Recipient{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last: "Mouse",
				Data: map[string]string {
					"ORG": "Disney",
				},
			},
		}
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].Email > expected[j].Email
		})
		sort.Slice(cfg.Recipients, func(i, j int) bool {
			return cfg.Recipients[i].Email > cfg.Recipients[j].Email
		})
		So(cfg.Recipients, ShouldResemble, expected)
	})
}

func TestLoadEmpty(t *testing.T) {
	Convey("Load sample config, test general parts", t, func() {
		_, err := New([]byte(`
[general]
`),
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "'mail-prog' not configured!")
	})
}

func TestLoadNoRecipients(t *testing.T) {
	Convey("Load sample config, test general parts", t, func() {
		_, err := New([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
`),
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "section 'recipients' does not exist")
	})
}

func TestLoadFull(t *testing.T) {
	Convey("Load full config, test general parts", t, func() {
		cfg, err := New([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
sender-email=rts@example.com
sender-name=Frodo Baggins
Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)

		So(cfg.MailProg, ShouldEqual, "gnu-mail")
		So(cfg.SenderEmail, ShouldEqual, "rts@example.com")
		So(cfg.SenderName, ShouldEqual, "Frodo Baggins")
		So(len(cfg.Cc), ShouldEqual, 2)
		So(cfg.Cc, ShouldResemble, []string{"weirdo@nsb.gov", "cc@example.com"})
		expected := []Recipient {
			Recipient{
				Email: "jd@example.com",
				First: "John",
				Last: "Doe Jr.",
				Data: map[string]string {
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			Recipient{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last: "Mouse",
				Data: map[string]string {
					"ORG": "Disney",
				},
			},
		}
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].Email > expected[j].Email
		})
		sort.Slice(cfg.Recipients, func(i, j int) bool {
			return cfg.Recipients[i].Email > cfg.Recipients[j].Email
		})
		So(cfg.Recipients, ShouldResemble, expected)
	})
}
func TestParseRecipients(t *testing.T) {
	Convey("Load the recipients", t, func() {
		cfg, err := ini.Load([]byte(`
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
sender-email=rts@example.com
sender-name=Frodo Baggins
Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)

		var recipients *ini.Section
		recipients, err = cfg.GetSection("recipients")
		So(err, ShouldBeNil)

		actual := parseRecipients(recipients)
		So(actual, ShouldNotBeNil)

		expected := []Recipient {
			Recipient{
				Email: "jd@example.com",
				First: "John",
				Last: "Doe Jr.",
				Data: map[string]string {
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			Recipient{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last: "Mouse",
				Data: map[string]string {
					"ORG": "Disney",
				},
			},
		}
		// sort expected / actual so different element ordering does not break
		// the test
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].Email > expected[j].Email
		})
		sort.Slice(actual, func(i, j int) bool {
			return actual[i].Email > actual[j].Email
		})
		So(actual, ShouldResemble, expected)
	})
}
