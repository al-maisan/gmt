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

package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMessage(t *testing.T) {
	msg := createMessage(
		`"Sender" <sender@example.com>`,
		"John Doe", "jd@example.com",
		nil, "", "Hello!", "Body text",
	)
	assert.NotNil(t, msg)
	assert.Equal(t, []string{"Hello!"}, msg.GetHeader("Subject"))
}

func TestCreateMessageWithCc(t *testing.T) {
	msg := createMessage(
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

func TestCreateMessageUTF8Name(t *testing.T) {
	msg := createMessage(
		`"Sender" <sender@example.com>`,
		"abc ähm", "abc@example.com",
		nil, "", "Hello!", "Body",
	)
	assert.NotNil(t, msg)
}

func TestAddAttachmentsNonexistent(t *testing.T) {
	msg := createMessage(
		`"Sender" <s@example.com>`,
		"A", "a@b.com",
		nil, "", "s", "b",
	)
	err := addAttachments(msg, []string{"/nonexistent/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment /nonexistent/file.txt")
}

func TestAddAttachmentsEmpty(t *testing.T) {
	msg := createMessage(
		`"Sender" <s@example.com>`,
		"A", "a@b.com",
		nil, "", "s", "b",
	)
	err := addAttachments(msg, nil)
	assert.NoError(t, err)
}

func TestLoadCredentialsMissing(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SENDER_EMAIL", "")
	t.Setenv("SENDER_PASSWORD", "")

	_, err := LoadCredentials()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP_HOST")
	assert.Contains(t, err.Error(), "SMTP_PORT")
	assert.Contains(t, err.Error(), "SENDER_EMAIL")
	assert.Contains(t, err.Error(), "SENDER_PASSWORD")
}

func TestLoadCredentialsPartialMissing(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SENDER_EMAIL", "")
	t.Setenv("SENDER_PASSWORD", "secret")

	_, err := LoadCredentials()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SENDER_EMAIL")
	assert.NotContains(t, err.Error(), "SMTP_HOST")
}

func TestLoadCredentialsBadPort(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "abc")
	t.Setenv("SENDER_EMAIL", "user@example.com")
	t.Setenv("SENDER_PASSWORD", "secret")

	_, err := LoadCredentials()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP_PORT")
	assert.Contains(t, err.Error(), "abc")
}

func TestLoadCredentialsValid(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SENDER_EMAIL", "user@example.com")
	t.Setenv("SENDER_PASSWORD", "secret")

	creds, err := LoadCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "smtp.example.com", creds.Host)
	assert.Equal(t, 587, creds.Port)
	assert.Equal(t, "user@example.com", creds.User)
	assert.Equal(t, "secret", creds.Password)
}
