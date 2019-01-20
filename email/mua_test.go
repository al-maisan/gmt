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
		args := PrepMUAArgs(cfg)

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-c", "'ab@cd.org, ef@gh.com, ij@kl.net'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSender(t *testing.T) {
	Convey("command line args, mailx [From]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			From:     "Hello Go <hello@go.go>",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-S", "from='Hello Go <hello@go.go>'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSenderAndNoName(t *testing.T) {
	Convey("command line args, mailx [From/p]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			From:     "hello@go.go",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-S", "from='hello@go.go'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithReplyTo(t *testing.T) {
	Convey("command line args, mailx [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-S", "replyto='Ja Mann <ja@mango.go>'"}

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-c", "'ab@cd.org, ef@gh.com, ij@kl.net'"}
		expected = append(expected, "-S", "from='Hello Go <hello@go.go>'")

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-c", "'ab@cd.org, ef@gh.com, ij@kl.net'"}
		expected = append(expected, "-S", "from='Hello Go <hello@go.go>'")
		expected = append(expected, "-S", "replyto='Ja Mann <ja@mango.go>'")

		So(args, ShouldResemble, expected)
	})
}

// ------------- non-mailx -------------

func TestPrepMUAArgsForNonMailxWithNoAdditionalData(t *testing.T) {
	Convey("minimal command line args for gnu-mail", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
		}
		args := PrepMUAArgs(cfg)

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithSender(t *testing.T) {
	Convey("command line args, gnu-mail [From]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			From:     "Hello Go <hello@go.go>",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'From: Hello Go <hello@go.go>'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithSenderAndNoName(t *testing.T) {
	Convey("command line args, gnu-mail [From/p]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			From:     "hello@go.go",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'From: hello@go.go'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForNonMailxWithReplyTo(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "gnu-mail",
			ReplyTo:  "Ja Mann <ja@mango.go>",
		}
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'Reply-To: Ja Mann <ja@mango.go>'"}

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		expected = append(expected, "-a", "'From: Hello Go <hello@go.go>'")

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
		args := PrepMUAArgs(cfg)

		expected := []string{cfg.MailProg, "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		expected = append(expected, "-a", "'From: Hello Go <hello@go.go>'")
		expected = append(expected, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")

		So(args, ShouldResemble, expected)
	})
}

func TestPostProcessMUAArgsWithNoCc(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		args := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		args = append(args, "-a", "'From: Hello Go <hello@go.go>'")
		args = append(args, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")

		data := Data{}
		actual := PostProcessMUAArgs(data, args)
		So(actual, ShouldResemble, args)
	})
}

func TestPostProcessMUAArgsWithCcRedefined(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		args := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		args = append(args, "-a", "'From: Hello Go <hello@go.go>'")
		args = append(args, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")

		data := Data{
			RecipientVars: map[string]string{
				"Cc": "rv1@redef.org, rv2@redef.com",
			},
		}
		expected := []string{"gnu-mail", "-a", "'Cc: rv1@redef.org, rv2@redef.com'"}
		expected = append(expected, "-a", "'From: Hello Go <hello@go.go>'")
		expected = append(expected, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")
		actual := PostProcessMUAArgs(data, args)
		So(actual, ShouldResemble, expected)
	})
}

func TestPostProcessMUAArgsWithSingleCcAdded(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		args := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		args = append(args, "-a", "'From: Hello Go <hello@go.go>'")
		args = append(args, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")

		data := Data{
			RecipientVars: map[string]string{
				"Cc": "+rv2@redef.com",
			},
		}
		expected := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net, rv2@redef.com'"}
		expected = append(expected, "-a", "'From: Hello Go <hello@go.go>'")
		expected = append(expected, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")
		actual := PostProcessMUAArgs(data, args)
		So(actual, ShouldResemble, expected)
	})
}

func TestPostProcessMUAArgsWithMultiCcAdded(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		args := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net'"}
		args = append(args, "-a", "'From: Hello Go <hello@go.go>'")
		args = append(args, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")

		data := Data{
			RecipientVars: map[string]string{
				"Cc": "+rv1@redef.org, rv2@redef.com",
			},
		}
		expected := []string{"gnu-mail", "-a", "'Cc: ab@cd.org, ef@gh.com, ij@kl.net, rv1@redef.org, rv2@redef.com'"}
		expected = append(expected, "-a", "'From: Hello Go <hello@go.go>'")
		expected = append(expected, "-a", "'Reply-To: Ja Mann <ja@mango.go>'")
		actual := PostProcessMUAArgs(data, args)
		So(actual, ShouldResemble, expected)
	})
}

func TestFindCcArgsMailxNoCcHeader(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"mailx"}
		So(findCcArgs(cmdline), ShouldEqual, -1)
	})
}

func TestFindCcArgsMailxNoCcHeaderBlah(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"mailx", "-S", "blah='yes or no?'", "-S", "herewego='Nver say never'"}
		So(findCcArgs(cmdline), ShouldEqual, -1)
	})
}

func TestFindCcArgsMailxCcHeaderBlah(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"mailx", "-S", "blah='yes or no?'", "-c", "whatever", "-S", "herewego='Nver say never'"}
		So(findCcArgs(cmdline), ShouldEqual, 4)
	})
}

func TestFindCcArgsNotMailxNoCcHeader(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"gnu-mail"}
		So(findCcArgs(cmdline), ShouldEqual, -1)
	})
}

func TestFindCcArgsNotMailxNoCcHeaderBlah(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"gnu-mail", "-a", "'From: whom?'", "-a", "'Reply-To: never!'"}
		So(findCcArgs(cmdline), ShouldEqual, -1)
	})
}

func TestFindCcArgsNotMailxCcHeaderBlah(t *testing.T) {
	Convey("command line args, gnu-mail [Reply-To, Cc, From]", t, func() {
		cmdline := []string{"gnu-mail", "-a", "'From: whom?'", "-a", "'Cc: carbon?'", "-a", "'Reply-To: never!'"}
		So(findCcArgs(cmdline), ShouldEqual, 4)
	})
}
