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

package email

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

func TestAttachFilesNonexistent(t *testing.T) {
	msg := createMessage(`"S" <s@s.com>`, "A", "a@b.com", nil, "", "s", "b")
	err := attachFiles(msg, []string{"/nonexistent/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment /nonexistent/file.txt")
}

func TestAttachFilesEmpty(t *testing.T) {
	msg := createMessage(`"S" <s@s.com>`, "A", "a@b.com", nil, "", "s", "b")
	assert.NoError(t, attachFiles(msg, nil))
}

func TestLoadSMTPCredentials(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "all missing",
			env:     map[string]string{"SMTP_HOST": "", "SMTP_PORT": "", "SENDER_EMAIL": "", "SENDER_PASSWORD": ""},
			wantErr: "SMTP_HOST",
		},
		{
			name:    "partial missing",
			env:     map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_PORT": "587", "SENDER_EMAIL": "", "SENDER_PASSWORD": "secret"},
			wantErr: "SENDER_EMAIL",
		},
		{
			name:    "bad port",
			env:     map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_PORT": "abc", "SENDER_EMAIL": "u@e.com", "SENDER_PASSWORD": "secret"},
			wantErr: "SMTP_PORT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			_, err := LoadSMTPCredentials()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadSMTPCredentialsValid(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SENDER_EMAIL", "user@example.com")
	t.Setenv("SENDER_PASSWORD", "secret")

	creds, err := LoadSMTPCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "smtp.example.com", creds.Host)
	assert.Equal(t, 587, creds.Port)
	assert.Equal(t, "user@example.com", creds.User)
	assert.Equal(t, "secret", creds.Password)
}
