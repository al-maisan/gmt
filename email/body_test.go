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
	"strconv"
	"testing"

	"github.com/al-maisan/gmt/config"
	. "github.com/smartystreets/goconvey/convey"
)

func GetConfig(input string) config.Data {
	cfg, _ := config.New([]byte(input))
	return cfg
}

func TestPrepBodies(t *testing.T) {
	Convey("generate email bodies for default config", t, func() {
		version := "0.2.0"
		cfg := GetConfig(config.SampleConfig(version))
		template := config.SampleTemplate(version)
		mails := PrepMails(cfg, template)

		expected := `FN / LN / EA = first name / last name / email address

Hello John // Doe Jr., how are things going at EFF?
this is your email: jd@example.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		mail := mails[0]
		So(mail.Recipient, ShouldEqual, "jd@example.com")
		So(mail.Body, ShouldEqual, expected)
		So(mail.Cmdline, ShouldEqual, []string{"gnu-mail", "-a", "Cc: bl@kf.io, info@ex.org", "-a", fmt.Sprintf("From: %s <rts@example.com>", strconv.Quote("Frodo Baggins")), "-s", "Hello John!", "jd@example.com"})

		expected = `FN / LN / EA = first name / last name / email address

Hello Mickey // Mouse, how are things going at Disney?
this is your email: mm@gmail.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		mail = mails[1]
		So(mail.Recipient, ShouldEqual, "mm@gmail.com")
		So(mail.Body, ShouldEqual, expected)
		So(mail.Cmdline, ShouldEqual, []string{})

		expected = `FN / LN / EA = first name / last name / email address

Hello Daisy // Lila, how are things going at NASA?
this is your email: daisy@example.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		mail = mails[2]
		So(mail.Recipient, ShouldEqual, "daisy@example.com")
		So(mail.Body, ShouldEqual, expected)
		So(mail.Cmdline, ShouldEqual, []string{})
	})
}

func TestPrepBodyForMailxWithNoAdditionalData(t *testing.T) {
	Convey("minimal command line args for mailx", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
		}
		recipient := config.Recipient{
			Email: "r1@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #1"

		expected := "body #1"
		body := prepBody(cfg, recipient, subject, "body #1")

		So(body, ShouldResemble, expected)
	})
}
