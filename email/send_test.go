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
	"strings"
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

func TestSendAllProgressAlignment(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := make([]Message, 12)
	for i := range msgs {
		msgs[i] = Message{Name: fmt.Sprintf("R%d", i+1), Address: fmt.Sprintf("r%d@example.com", i+1), Subject: "Hi", Body: "Hello"}
	}

	var buf bytes.Buffer
	NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
	lines := strings.Split(buf.String(), "\n")

	// First and last progress prefixes should be the same width
	assert.Contains(t, lines[0], "[ 1/12]")
	assert.Contains(t, lines[11], "[12/12]")

	// All prefix widths should match
	for i := 0; i < 12; i++ {
		assert.Contains(t, lines[i], "/12]")
	}
}

func TestSendAllSuccess(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Jane", Address: "jane@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
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
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
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
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
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
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, buf.String(), "failed to attach")
}

func TestSendAllMixedResults(t *testing.T) {
	sender := &failNthSender{failOn: 2}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Jane", Address: "jane@example.com", Subject: "Hi", Body: "Hello"},
		{Name: "Bob", Address: "bob@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
	assert.Equal(t, 2, result.Sent)
	assert.Equal(t, 1, result.Failed)
}

// failNthSender fails on the Nth Send call, succeeds on all others.
type failNthSender struct {
	calls  int
	failOn int
}

func (f *failNthSender) Send(_ *mail.Msg) error {
	f.calls++
	if f.calls == f.failOn {
		return fmt.Errorf("simulated failure on send #%d", f.failOn)
	}
	return nil
}

func (f *failNthSender) Close() error { return nil }

func TestSendAllRetrySuccess(t *testing.T) {
	// Fails on first call, succeeds on second (the retry)
	sender := &failNthSender{failOn: 1}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := NewBatchSender(&buf, sender, cfg, SendOptions{Retries: 1}).SendAll(msgs)
	assert.Equal(t, 1, result.Sent)
	assert.Equal(t, 0, result.Failed)
	assert.Contains(t, buf.String(), "retrying")
	assert.Contains(t, buf.String(), "- John <jd@example.com>")
}

func TestSendAllRetryExhausted(t *testing.T) {
	// Always fails — retries should be exhausted
	sender := &mockSender{sendErr: fmt.Errorf("permanent error")}
	cfg := config.MailConfig{From: "sender@example.com"}
	msgs := []Message{
		{Name: "John", Address: "jd@example.com", Subject: "Hi", Body: "Hello"},
	}

	var buf bytes.Buffer
	result := NewBatchSender(&buf, sender, cfg, SendOptions{Retries: 2}).SendAll(msgs)
	assert.Equal(t, 0, result.Sent)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 2, strings.Count(buf.String(), "retrying"))
	assert.Contains(t, buf.String(), "permanent error")
}

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
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
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
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(msgs)
	assert.Equal(t, 1, result.Sent)
	assert.Equal(t, 0, result.Failed)
}

func TestSendAllEmpty(t *testing.T) {
	sender := &mockSender{}
	cfg := config.MailConfig{From: "sender@example.com"}

	var buf bytes.Buffer
	result := NewBatchSender(&buf, sender, cfg, SendOptions{}).SendAll(nil)
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
