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
	"sort"
	"testing"

	"github.com/go-ini/ini"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadDefault(t *testing.T) {
	Convey("Load sample config, test general parts", t, func(c C) {
		cfg, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
#Cc=weirdo@nsb.gov, cc@example.com
#Reply-to=John Doe <jd@mail.com>
subject=Hello %FN%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		c.So(err, ShouldBeNil)
		c.So(cfg, ShouldNotBeNil)

		c.So(cfg.From, ShouldEqual, "Frodo Baggins <rts@example.com>")
		c.So(len(cfg.Cc), ShouldEqual, 0)
		expected := []Recipient{
			{
				Email: "jd@example.com",
				First: "John",
				Last:  "Doe Jr.",
				Data: map[string]string{
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last:  "Mouse",
				Data: map[string]string{
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
		c.So(cfg.Recipients, ShouldResemble, expected)
	})
}

func TestLoadNoRecipients(t *testing.T) {
	Convey("Load sample config missing a 'recipients' section", t, func(c C) {
		_, err := New([]byte(`
[general]
subject=Hello %FN%!
`),
		)
		c.So(err, ShouldNotBeNil)
		c.So(err.Error(), ShouldContainSubstring, "recipients")
	})
}

func TestLoadNoSubject(t *testing.T) {
	Convey("Load sample config missing a subject definition", t, func(c C) {
		_, err := New([]byte(`
[general]
`),
		)
		c.So(err, ShouldNotBeNil)
		c.So(err.Error(), ShouldEqual, "'subject' not configured!")
	})
}

func TestLoadSubjectCaseInsensitive(t *testing.T) {
	Convey("Load config with title-case Subject key", t, func(c C) {
		cfg, err := New([]byte(`
[general]
Subject=Hello %FN%!
[recipients]
jd@example.com=John Doe
`),
		)
		c.So(err, ShouldBeNil)
		c.So(cfg.Subject, ShouldEqual, "Hello %FN%!")
	})
}

func TestLoadFull(t *testing.T) {
	Convey("Load full config, test general parts", t, func(c C) {
		cfg, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
Cc=weirdo@nsb.gov, cc@example.com
Reply-To=John Doe <jd@mail.com>
subject=Hello %FN%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		c.So(err, ShouldBeNil)
		c.So(cfg, ShouldNotBeNil)

		c.So(cfg.From, ShouldEqual, "Frodo Baggins <rts@example.com>")
		c.So(cfg.ReplyTo, ShouldEqual, "John Doe <jd@mail.com>")
		c.So(len(cfg.Cc), ShouldEqual, 2)
		c.So(cfg.Subject, ShouldEqual, "Hello %FN%!")
		c.So(cfg.Cc, ShouldResemble, []string{"weirdo@nsb.gov", "cc@example.com"})
		expected := []Recipient{
			{
				Email: "jd@example.com",
				First: "John",
				Last:  "Doe Jr.",
				Data: map[string]string{
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last:  "Mouse",
				Data: map[string]string{
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
		c.So(cfg.Recipients, ShouldResemble, expected)
	})
}

func TestParseRecipients(t *testing.T) {
	Convey("Load the recipients", t, func(c C) {
		cfg, err := ini.InsensitiveLoad([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`),
		)
		c.So(err, ShouldBeNil)
		c.So(cfg, ShouldNotBeNil)

		var recipients *ini.Section
		recipients, err = cfg.GetSection("recipients")
		c.So(err, ShouldBeNil)

		actual := parseRecipients(recipients)
		c.So(actual, ShouldNotBeNil)

		expected := []Recipient{
			{
				Email: "jd@example.com",
				First: "John",
				Last:  "Doe Jr.",
				Data: map[string]string{
					"ORG": "EFF", "TITLE": "PhD",
				},
			},
			{
				Email: "mm@gmail.com",
				First: "Mickey",
				Last:  "Mouse",
				Data: map[string]string{
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
		c.So(actual, ShouldResemble, expected)
	})
}

func TestParseRecipientsSingleName(t *testing.T) {
	Convey("Parse a recipient with only a first name", t, func(c C) {
		cfg, err := ini.InsensitiveLoad([]byte(`
[recipients]
madonna@example.com=Madonna
`),
		)
		c.So(err, ShouldBeNil)

		recipients, err := cfg.GetSection("recipients")
		c.So(err, ShouldBeNil)

		actual := parseRecipients(recipients)
		c.So(len(actual), ShouldEqual, 1)
		c.So(actual[0].First, ShouldEqual, "Madonna")
		c.So(actual[0].Last, ShouldEqual, "")
		c.So(actual[0].Email, ShouldEqual, "madonna@example.com")
	})
}

func TestParseRecipientsMalformedData(t *testing.T) {
	Convey("Parse a recipient with malformed custom data (missing :-)", t, func(c C) {
		cfg, err := ini.InsensitiveLoad([]byte(`
[recipients]
jd@example.com=John Doe|BADDATA|ORG:-EFF
`),
		)
		c.So(err, ShouldBeNil)

		recipients, err := cfg.GetSection("recipients")
		c.So(err, ShouldBeNil)

		actual := parseRecipients(recipients)
		c.So(len(actual), ShouldEqual, 1)
		// BADDATA should be skipped, ORG should be parsed
		c.So(actual[0].Data, ShouldResemble, map[string]string{"ORG": "EFF"})
	})
}
