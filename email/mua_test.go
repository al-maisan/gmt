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

		So(len(args), ShouldEqual, 0)
		expected := []string{}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithCc(t *testing.T) {
	Convey("command line args, mailx [Cc]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc: []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-c",  "ab@cd.org, ef@gh.com, ij@kl.net"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSender(t *testing.T) {
	Convey("command line args, mailx [Sender*]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			SenderName: "Hello Go",
			SenderEmail: "hello@go.go",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-r", "Hello Go <hello@go.go>"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSenderAndNoName(t *testing.T) {
	Convey("command line args, mailx [Sender*]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			SenderEmail: "hello@go.go",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-r", "hello@go.go"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithSenderAndNoEmail(t *testing.T) {
	Convey("command line args, mailx [SenderName]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			SenderName: "Hello Go",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithReplyTo(t *testing.T) {
	Convey("command line args, mailx [Reply-To]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			ReplyTo: "Ja Mann <ja@mango.go>",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-S", "replyto='Ja Mann <ja@mango.go>'"}

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithCcAndSender(t *testing.T) {
	Convey("command line args, mailx [Cc, Sender*]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc: []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			SenderName: "Hello Go",
			SenderEmail: "hello@go.go",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-c", "ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, []string{"-r", "Hello Go <hello@go.go>"}...)

		So(args, ShouldResemble, expected)
	})
}

func TestPrepMUAArgsForMailxWithAll(t *testing.T) {
	Convey("command line args, mailx [Reply-To, Cc, Sender*]", t, func() {
		cfg := config.Data{
			MailProg: "mailx",
			Cc: []string{"ab@cd.org", "ef@gh.com", "ij@kl.net"},
			SenderName: "Hello Go",
			SenderEmail: "hello@go.go",
			ReplyTo: "Ja Mann <ja@mango.go>",
			Subject: "This is spam!",
		}
		args := PrepMUAArgs(cfg)

		// So(len(args), ShouldEqual, 1)
		expected := []string{"-c", "ab@cd.org, ef@gh.com, ij@kl.net"}
		expected = append(expected, []string{"-r", "Hello Go <hello@go.go>"}...)
		expected = append(expected, []string{"-S", "replyto='Ja Mann <ja@mango.go>'"}...)
		expected = append(expected, []string{"-s", "This is spam!"}...)

		So(args, ShouldResemble, expected)
	})
}
