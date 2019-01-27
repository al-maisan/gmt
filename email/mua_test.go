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

func TestPrepMUAArgsForMailxWithNoAdditionalData(t *testing.T) {
	Convey("minimal command line args for mailx", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		So(len(args), ShouldEqual, 1)
		expected := []string{cfg.MailProg}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithCc(t *testing.T) {
	Convey("command line args, mailx [Cc]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg}
		expected = append(expected, "-c", "ab@cd.org,ef@gh.com,ij@kl.net")

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSender(t *testing.T) {
	Convey("command line args, mailx [From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-S", "from=Hello Go <hello@go.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSenderAndNoName(t *testing.T) {
	Convey("command line args, mailx [From/p]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			From:     "hello@go.go",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-S", "from=hello@go.go"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithReplyTo(t *testing.T) {
	Convey("command line args, mailx [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-S", "replyto=Ja Mann <ja@mango.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithCcAndSender(t *testing.T) {
	Convey("command line args, mailx [Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg}
		expected = append(expected, "-c", "ab@cd.org,ef@gh.com,ij@kl.net")
		expected = append(expected, "-S", "from=Hello Go <hello@go.go>")

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithAll(t *testing.T) {
	Convey("command line args, mailx [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "This is spam!",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg}
		expected = append(expected, "-c", "ab@cd.org,ef@gh.com,ij@kl.net")
		expected = append(expected, "-S", "from=Hello Go <hello@go.go>")
		expected = append(expected, "-S", "replyto=Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it redefines/overrides the global `Cc`
// header value
func TestPrepMUAArgsForMailxWithAllAndPRCcRedef(t *testing.T) {
	Convey("command line args, mailx [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "This is spam!",
		}
		prdata := map[string]string{
			"Cc": "hello1@world.com,   	2nd@example.org",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "-c", "hello1@world.com,2nd@example.org")
		expected = append(expected, "-S", "from=Hello Go <hello@go.go>")
		expected = append(expected, "-S", "replyto=Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it adds to the global `Cc`
// header value
func TestPrepMUAArgsForMailxWithAllAndPRCcAdded(t *testing.T) {
	Convey("command line args, mailx [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "This is spam!",
		}
		prdata := map[string]string{
			"Cc": "+am@world.com,   	mtp@example.org",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "-c", "ab@cd.org,ef@gh.com,ij@kl.net,am@world.com,mtp@example.org")
		expected = append(expected, "-S", "from=Hello Go <hello@go.go>")
		expected = append(expected, "-S", "replyto=Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// ------------- non-mailx -------------

func TestPrepMUAArgsForNonMailxWithNoAdditionalData(t *testing.T) {
	Convey("minimal command line args for gnu-mail", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		So(len(args), ShouldEqual, 1)
		expected := []string{cfg.MailProg}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithCc(t *testing.T) {
	Convey("command line args, gnu-mail [Cc]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "Cc: ab@cd.org, ef@gh.com, ij@kl.net"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithSender(t *testing.T) {
	Convey("command line args, gnu-mail [From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "From: Hello Go <hello@go.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithSenderAndNoName(t *testing.T) {
	Convey("command line args, gnu-mail [From/p]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			From:     "hello@go.go",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "From: hello@go.go"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithReplyTo(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "Reply-To: Ja Mann <ja@mango.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithCcAndSender(t *testing.T) {
	Convey("command line args, gnu-mail [Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "Cc: ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, "-a", "From: Hello Go <hello@go.go>")

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithAll(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "-a", "Cc: ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, "-a", "From: Hello Go <hello@go.go>")
		expected = append(expected, "-a", "Reply-To: Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it redefines/overrides the global `Cc`
// header value
func TestPrepMUAArgsForNonMailxWithAllAndPRCcRedef(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		prdata := map[string]string{
			"Cc": "First One <hello1@world.com>,   	The Second <2nd@example.org>",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "-a", "Cc: First One <hello1@world.com>, The Second <2nd@example.org>")
		expected = append(expected, "-a", "From: Hello Go <hello@go.go>")
		expected = append(expected, "-a", "Reply-To: Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it adds to the global `Cc`
// header value
func TestPrepMUAArgsForNonMailxWithAllAndPRCcAdded(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		prdata := map[string]string{
			"Cc": "+Add Me <am@world.com>,   	Me Too Please Second <mtp@example.org>",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "-a", "Cc: ab@cd.org, ef@gh.com, ij@kl.net, Add Me <am@world.com>, Me Too Please Second <mtp@example.org>")
		expected = append(expected, "-a", "From: Hello Go <hello@go.go>")
		expected = append(expected, "-a", "Reply-To: Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// ------------- sendmail --------------

func TestPrepMUAArgsForSendmailWithNoAdditionalData(t *testing.T) {
	Convey("minimal command line args for sendmail", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		So(len(args), ShouldEqual, 1)
		expected := []string{cfg.MailProg}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithCc(t *testing.T) {
	Convey("command line args, sendmail [Cc]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "Cc:", "ab@cd.org, ef@gh.com, ij@kl.net"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithSender(t *testing.T) {
	Convey("command line args, sendmail [From]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "From:", "Hello Go <hello@go.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithSenderAndNoName(t *testing.T) {
	Convey("command line args, sendmail [From/p]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			From:     "hello@go.go",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "From:", "hello@go.go"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithReplyTo(t *testing.T) {
	Convey("command line args, sendmail [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "Reply-To:", "Ja Mann <ja@mango.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithCcAndSender(t *testing.T) {
	Convey("command line args, sendmail [Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "Cc:", "ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, "From:", "Hello Go <hello@go.go>")

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForSendmailWithAll(t *testing.T) {
	Convey("command line args, sendmail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		args := prepMUAArgs(cfg, map[string]string{})

		expected := []string{cfg.MailProg, "Cc:", "ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, "From:", "Hello Go <hello@go.go>")
		expected = append(expected, "Reply-To:", "Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it redefines/overrides the global `Cc`
// header value
func TestPrepMUAArgsForSendmailWithAllAndPRCcRedef(t *testing.T) {
	Convey("command line args, sendmail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		prdata := map[string]string{
			"Cc": "First One <hello1@world.com>,   	The Second <2nd@example.org>",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "Cc:", "First One <hello1@world.com>, The Second <2nd@example.org>")
		expected = append(expected, "From:", "Hello Go <hello@go.go>")
		expected = append(expected, "Reply-To:", "Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}

// A `Cc` is set for the recipient and it adds to the global `Cc`
// header value
func TestPrepMUAArgsForSendmailWithAllAndPRCcAdded(t *testing.T) {
	Convey("command line args, sendmail [Reply-To, Cc, From]", t, func() {
		cfg := config.Data{
			MailProg: "sendmail",
			Cc:       []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			From:     "Hello Go <hello@go.go>",
			ReplyTo:  "Ja Mann <ja@mango.go>",
			Subject:  "Hola %FN%!",
		}
		prdata := map[string]string{
			"Cc": "+Add Me <am@world.com>,   	Me Too Please Second <mtp@example.org>",
		}
		args := prepMUAArgs(cfg, prdata)

		expected := []string{cfg.MailProg}
		expected = append(expected, "Cc:", "ab@cd.org, ef@gh.com, ij@kl.net, Add Me <am@world.com>, Me Too Please Second <mtp@example.org>")
		expected = append(expected, "From:", "Hello Go <hello@go.go>")
		expected = append(expected, "Reply-To:", "Ja Mann <ja@mango.go>")

		So(args, ShouldResemble, expected)
	})
}
