// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019-2025  "Muharem Hrnjadovic" <muharem@linux.com>
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

	"github.com/stretchr/testify/assert"
)

func TestCreateEmailMessage(t *testing.T) {
	msg := createEmailMessage(
		`"Sender" <sender@example.com>`,
		"John Doe", "jd@example.com",
		nil, "", "Hello!", "Body text",
	)
	assert.NotNil(t, msg)
	assert.Equal(t, []string{"Hello!"}, msg.GetHeader("Subject"))
}

func TestCreateEmailMessageWithCc(t *testing.T) {
	msg := createEmailMessage(
		`"Sender" <sender@example.com>`,
		"John Doe", "jd@example.com",
		[]string{"cc1@example.com", "cc2@example.com"},
		"reply@example.com",
		"Test", "Body",
	)
	assert.NotNil(t, msg)
	assert.Equal(t, []string{"cc1@example.com,cc2@example.com"}, msg.GetHeader("Cc"))
	assert.Equal(t, []string{"reply@example.com"}, msg.GetHeader("Reply-To"))
}

func TestCreateEmailMessageUTF8Name(t *testing.T) {
	msg := createEmailMessage(
		`"Sender" <sender@example.com>`,
		"abc ähm", "abc@example.com",
		nil, "", "Hello!", "Body",
	)
	assert.NotNil(t, msg)
}

func TestAddAttachmentsNonexistent(t *testing.T) {
	msg := createEmailMessage(
		`"Sender" <s@example.com>`,
		"A", "a@b.com",
		nil, "", "s", "b",
	)
	err := addAttachments(msg, []string{"/nonexistent/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment /nonexistent/file.txt")
}

func TestAddAttachmentsEmpty(t *testing.T) {
	msg := createEmailMessage(
		`"Sender" <s@example.com>`,
		"A", "a@b.com",
		nil, "", "s", "b",
	)
	err := addAttachments(msg, nil)
	assert.NoError(t, err)
}
