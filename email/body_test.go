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
	"testing"

	"github.com/al-maisan/gmt/config"
	. "github.com/smartystreets/goconvey/convey"
)

// ------------- non-sendmail -------------

func TestPrepBodyForMailxWithNoAdditionalData(t *testing.T) {
	Convey("body for mailx", t, func(c C) {
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
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForMailxWithCc(t *testing.T) {
	Convey("body, mailx [Cc]", t, func(c C) {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		recipient := config.Recipient{
			Email: "r2@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! How are things? #2"

		expected := "body #2"
		body := prepBody(cfg, recipient, subject, "body #2")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForMailxWithSender(t *testing.T) {
	Convey("body, mailx [From]", t, func(c C) {
		cfg := config.Data{
			MailProg: "mailx",
			From:     "Hello Go <hello@go.go>",
		}
		recipient := config.Recipient{
			Email: "r3@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! How are things? #3"

		expected := "body #3"
		body := prepBody(cfg, recipient, subject, "body #3")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForMailxWithReplyTo(t *testing.T) {
	Convey("body, mailx [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "mailx",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		recipient := config.Recipient{
			Email: "r4@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! How are things? #4"

		expected := "body #4"
		body := prepBody(cfg, recipient, subject, "body #4")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForMailxWithCcAndSender(t *testing.T) {
	Convey("body, mailx [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
		}
		recipient := config.Recipient{
			Email: "r5@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! How are things? #5"

		expected := "body #5"
		body := prepBody(cfg, recipient, subject, "body #5")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForMailxWithAll(t *testing.T) {
	Convey("body, mailx [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "This is spam!",
		}
		recipient := config.Recipient{
			Email: "r6@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! How are things? #6"

		expected := "body #6"
		body := prepBody(cfg, recipient, subject, "body #6")
		c.So(body, ShouldEqual, expected)
	})
}

// ------------- sendmail -------------

func TestPrepBodyForSendmailWithNoAdditionalData(t *testing.T) {
	Convey("body for sendmail", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.7",
		}
		recipient := config.Recipient{
			Email: "r7@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #7"

		expected := `To: r7@example.com
Subject: Hello! Just throw it back #7
X-Mailer: gmt, version 0.99.7, https://301.mx/gmt

body #7`
		body := prepBody(cfg, recipient, subject, "body #7")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForSendmailWithCc(t *testing.T) {
	Convey("body, sendmail [Cc]", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.8",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		recipient := config.Recipient{
			Email: "r8@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #8"

		expected := `To: r8@example.com
Subject: Hello! Just throw it back #8
Cc: ab@cd.org, ef@gh.com, ij@kl.net
X-Mailer: gmt, version 0.99.8, https://301.mx/gmt

body #8`
		body := prepBody(cfg, recipient, subject, "body #8")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForSendmailWithSender(t *testing.T) {
	Convey("body, sendmail [From]", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.9",
			From:     "Hello Go <hello@go.go>",
		}
		recipient := config.Recipient{
			Email: "r9@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #9"

		expected := `To: r9@example.com
Subject: Hello! Just throw it back #9
From: Hello Go <hello@go.go>
X-Mailer: gmt, version 0.99.9, https://301.mx/gmt

body #9`
		body := prepBody(cfg, recipient, subject, "body #9")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForSendmailWithReplyTo(t *testing.T) {
	Convey("body, sendmail [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.A",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		recipient := config.Recipient{
			Email: "rA@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #A"

		expected := `To: rA@example.com
Subject: Hello! Just throw it back #A
Reply-To: Ja Mann <ja@mango.go>
X-Mailer: gmt, version 0.99.A, https://301.mx/gmt

body #A`
		body := prepBody(cfg, recipient, subject, "body #A")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForSendmailWithCcAndSender(t *testing.T) {
	Convey("body, sendmail [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.B",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
		}
		recipient := config.Recipient{
			Email: "rB@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #B"

		expected := `To: rB@example.com
Subject: Hello! Just throw it back #B
Cc: ab@cd.org, ef@gh.com, ij@kl.net
From: Hello Go <hello@go.go>
X-Mailer: gmt, version 0.99.B, https://301.mx/gmt

body #B`
		body := prepBody(cfg, recipient, subject, "body #B")
		c.So(body, ShouldEqual, expected)
	})
}

func TestPrepBodyForSendmailWithAll(t *testing.T) {
	Convey("body, sendmail [Reply-To]", t, func(c C) {
		cfg := config.Data{
			MailProg: "sendmail",
			Version:  "0.99.C",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "This is spam!",
		}
		recipient := config.Recipient{
			Email: "rC@example.com",
			First: "The",
			Last:  "Mailinator",
		}
		subject := "Hello! Just throw it back #C"

		expected := `To: rC@example.com
Subject: Hello! Just throw it back #C
Cc: ab@cd.org, ef@gh.com, ij@kl.net
From: Hello Go <hello@go.go>
Reply-To: Ja Mann <ja@mango.go>
X-Mailer: gmt, version 0.99.C, https://301.mx/gmt

body #C`
		body := prepBody(cfg, recipient, subject, "body #C")
		c.So(body, ShouldEqual, expected)
	})
}
