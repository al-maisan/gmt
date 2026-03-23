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
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/al-maisan/gmt/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mail "github.com/wneessen/go-mail"
)

func TestCreateMessage(t *testing.T) {
	msg, err := createMessage(
		"sender@example.com",
		"John Doe", "jd@example.com",
		nil, "", "Hello!", "Body text",
	)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

func TestCreateMessageWithCcAndReplyTo(t *testing.T) {
	msg, err := createMessage(
		"sender@example.com",
		"John Doe", "jd@example.com",
		[]string{"cc1@example.com", "cc2@example.com"},
		"reply@example.com",
		"Test", "Body",
	)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

func TestCreateMessageUTF8Name(t *testing.T) {
	msg, err := createMessage(
		"sender@example.com",
		"abc ähm", "abc@example.com",
		nil, "", "Hello!", "Body",
	)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

func TestCreateMessageInvalidFrom(t *testing.T) {
	_, err := createMessage(
		"not-an-email",
		"A", "a@b.com",
		nil, "", "s", "b",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid From")
}

func TestCreateMessageInvalidTo(t *testing.T) {
	_, err := createMessage(
		"sender@example.com",
		"A", "not-an-email",
		nil, "", "s", "b",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid To")
}

func TestAttachFilesNonexistent(t *testing.T) {
	msg, err := createMessage("s@s.com", "A", "a@b.com", nil, "", "s", "b")
	require.NoError(t, err)
	err = attachFiles(msg, []string{"/nonexistent/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment /nonexistent/file.txt")
}

func TestAttachFilesEmpty(t *testing.T) {
	msg, err := createMessage("s@s.com", "A", "a@b.com", nil, "", "s", "b")
	require.NoError(t, err)
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

func TestCreateMessageWithReplyTo(t *testing.T) {
	msg, err := createMessage(
		"sender@example.com",
		"John Doe", "jd@example.com",
		nil,
		`"Support" <support@example.com>`,
		"Test", "Body",
	)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

func TestCreateMessageInvalidCc(t *testing.T) {
	_, err := createMessage(
		"sender@example.com",
		"A", "a@b.com",
		[]string{"not-an-email"}, "",
		"s", "b",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Cc")
}

func TestCreateMessageInvalidReplyTo(t *testing.T) {
	_, err := createMessage(
		"sender@example.com",
		"A", "a@b.com",
		nil, "not-an-email",
		"s", "b",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Reply-To")
}

func TestAttachFilesValid(t *testing.T) {
	msg, err := createMessage("s@s.com", "A", "a@b.com", nil, "", "s", "b")
	require.NoError(t, err)

	tmpFile := t.TempDir() + "/test.txt"
	require.NoError(t, os.WriteFile(tmpFile, []byte("content"), 0o644))
	assert.NoError(t, attachFiles(msg, []string{tmpFile}))
}

// mockSender records Send calls and returns a configurable error.
type mockSender struct {
	sent    int
	sendErr error
	closed  bool
}

func (m *mockSender) Send(_ *mail.Msg) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent++
	return nil
}

func (m *mockSender) Close() error {
	m.closed = true
	return nil
}

func TestNewSMTPSenderConnectionError(t *testing.T) {
	creds := SMTPCredentials{
		Host:     "localhost",
		Port:     19999,
		User:     "user",
		Password: "pass",
	}
	_, err := NewSMTPSender(creds)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestSendAllSuccess(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Jane", Address: "jane@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 2, result.Sent)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 2, sender.sent)
	assert.Contains(t, buf.String(), "- John <jd@example.com>")
	assert.Contains(t, buf.String(), "- Jane <jane@example.com>")
}

func TestSendAllSendError(t *testing.T) {
	sender := &mockSender{sendErr: fmt.Errorf("connection reset")}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, buf.String(), "! John <jd@example.com>")
	assert.Contains(t, buf.String(), "connection reset")
}

func TestSendAllInvalidFrom(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "not-an-email"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, buf.String(), "failed to create")
	assert.Equal(t, 0, sender.sent)
}

func TestSendAllMissingAttachment(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello", Attachments: []string{"/nonexistent/file.txt"}},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, buf.String(), "failed to attach")
}

func TestSendAllMixedResults(t *testing.T) {
	sender := &failNthSender{failMsg: 2}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Jane", Address: "jane@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Bob", Address: "bob@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 2, result.Sent)
	assert.Equal(t, 1, result.Failed)
}

// failNthSender permanently fails for message N (counting from 1).
// Since SendAll retries once, the sender must fail on both the original
// and retry calls to produce an actual failure.
type failNthSender struct {
	msgIndex  int // incremented only on first attempt (not retries)
	failMsg   int // which message number to fail (1-based)
	lastFailed bool
}

func (f *failNthSender) Send(_ *mail.Msg) error {
	if f.lastFailed {
		// This is a retry of the previously failed message.
		f.lastFailed = false
		return fmt.Errorf("simulated permanent failure on message #%d", f.failMsg)
	}
	f.msgIndex++
	if f.msgIndex == f.failMsg {
		f.lastFailed = true
		return fmt.Errorf("simulated failure on message #%d", f.failMsg)
	}
	return nil
}

func (f *failNthSender) Close() error { return nil }

func TestSendAllWithReplyTo(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{
		From:    "sender@example.com",
		ReplyTo: "reply@example.com",
	}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 1, result.Sent)
	assert.Equal(t, 0, result.Failed)
}

func TestSendAllWithCc(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello", Cc: []string{"cc@example.com"}},
	}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, msgs, 0)
	assert.Equal(t, 1, result.Sent)
	assert.Equal(t, 0, result.Failed)
}

func TestSendAllEmpty(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}

	var buf bytes.Buffer
	result := SendAll(&buf, sender, cfg, nil, 0)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, buf.String())
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
