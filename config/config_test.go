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

package config

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sortRecipients(r []Recipient) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].Email > r[j].Email
	})
}

func parseTestConfig(t *testing.T, input []byte) MailConfig {
	t.Helper()
	cfg, err := Parse(input)
	require.NoError(t, err)
	return cfg
}

func TestParseDefault(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "Frodo Baggins <rts@example.com>"
subject = "Hello %FN%!"

[[recipients]]
email = "jd@example.com"
first = "John"
last = "Doe Jr."
data = { ORG = "EFF", TITLE = "PhD" }

[[recipients]]
email = "mm@gmail.com"
first = "Mickey"
last = "Mouse"
data = { ORG = "Disney" }
`))

	assert.Equal(t, "Frodo Baggins <rts@example.com>", cfg.From)
	assert.Empty(t, cfg.Cc)

	expected := []Recipient{
		{Email: "jd@example.com", First: "John", Last: "Doe Jr.", Data: map[string]string{"ORG": "EFF", "TITLE": "PhD"}},
		{Email: "mm@gmail.com", First: "Mickey", Last: "Mouse", Data: map[string]string{"ORG": "Disney"}},
	}
	sortRecipients(expected)
	sortRecipients(cfg.Recipients)
	assert.Equal(t, expected, cfg.Recipients)
}

func TestParseMissingFrom(t *testing.T) {
	_, err := Parse([]byte(`
[general]
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required key 'from'")
}

func TestParseMissingSubject(t *testing.T) {
	_, err := Parse([]byte(`
[general]
from = "x <x@x.com>"
[[recipients]]
email = "a@b.com"
first = "A"
`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required key 'subject'")
}

func TestParseNoRecipients(t *testing.T) {
	_, err := Parse([]byte(`
[general]
from = "Frodo Baggins <rts@example.com>"
subject = "Hello %FN%!"
`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no [[recipients]] entries found")
}

func TestParseInvalidTOML(t *testing.T) {
	_, err := Parse([]byte(`this is not valid TOML {{{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TOML syntax error")
}

func TestParseFull(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "Frodo Baggins <rts@example.com>"
cc = ["weirdo@nsb.gov", "cc@example.com"]
reply_to = "John Doe <jd@mail.com>"
subject = "Hello %FN%!"

[[recipients]]
email = "jd@example.com"
first = "John"
last = "Doe Jr."
data = { ORG = "EFF", TITLE = "PhD" }

[[recipients]]
email = "mm@gmail.com"
first = "Mickey"
last = "Mouse"
data = { ORG = "Disney" }
`))

	assert.Equal(t, "Frodo Baggins <rts@example.com>", cfg.From)
	assert.Equal(t, "John Doe <jd@mail.com>", cfg.ReplyTo)
	assert.Equal(t, "Hello %FN%!", cfg.Subject)
	assert.Equal(t, []string{"weirdo@nsb.gov", "cc@example.com"}, cfg.Cc)

	expected := []Recipient{
		{Email: "jd@example.com", First: "John", Last: "Doe Jr.", Data: map[string]string{"ORG": "EFF", "TITLE": "PhD"}},
		{Email: "mm@gmail.com", First: "Mickey", Last: "Mouse", Data: map[string]string{"ORG": "Disney"}},
	}
	sortRecipients(expected)
	sortRecipients(cfg.Recipients)
	assert.Equal(t, expected, cfg.Recipients)
}

func TestParseRecipientSingleName(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "madonna@example.com"
first = "Madonna"
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "Madonna", cfg.Recipients[0].First)
	assert.Equal(t, "", cfg.Recipients[0].Last)
	assert.Equal(t, "madonna@example.com", cfg.Recipients[0].Email)
}

func TestParseRecipientMissingEmail(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
first = "John"
last = "Doe"
[[recipients]]
email = "valid@example.com"
first = "Valid"
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "valid@example.com", cfg.Recipients[0].Email)
	require.Len(t, cfg.Warnings, 1)
	assert.Contains(t, cfg.Warnings[0], "missing 'email'")
}

func TestParseRecipientMissingFirst(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "jd@example.com"
last = "Doe"
[[recipients]]
email = "valid@example.com"
first = "Valid"
`))
	require.Len(t, cfg.Recipients, 1)
	require.Len(t, cfg.Warnings, 1)
	assert.Contains(t, cfg.Warnings[0], "missing 'first'")
}

func TestParseRecipientCcReplace(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
cc = ["override@cc.com"]
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "override@cc.com", cfg.Recipients[0].Data["CC"])
}

func TestParseRecipientCcAppend(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
cc_extra = ["extra@cc.com"]
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "+extra@cc.com", cfg.Recipients[0].Data["CC"])
}

func TestParseRecipientAttachReplace(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
attachments = ["local.txt"]
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "local.txt", cfg.Recipients[0].Data["AS"])
}

func TestParseRecipientAttachAppend(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
attachments_extra = ["extra.pdf"]
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "+extra.pdf", cfg.Recipients[0].Data["AS"])
}

func TestParseDataKeysUppercased(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
from = "test <t@example.com>"
subject = "test"
[[recipients]]
email = "a@b.com"
first = "A"
data = { org = "EFF" }
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "EFF", cfg.Recipients[0].Data["ORG"])
}

func TestCheckAttachmentsValid(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	require.NoError(t, os.WriteFile(tmpFile, []byte("content"), 0o644))
	assert.NoError(t, checkAttachments([]string{tmpFile}))
}

func TestCheckAttachmentsEmpty(t *testing.T) {
	assert.NoError(t, checkAttachments(nil))
}

func TestCheckAttachmentsMissing(t *testing.T) {
	err := checkAttachments([]string{"/nonexistent/file.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment not found")
}

func TestCheckAttachmentsNotAccessible(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/noperm.txt"
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	// On most systems os.Stat succeeds even without read permission,
	// so this test just verifies the function doesn't panic.
	_ = checkAttachments([]string{path})
}

func TestParseWithAttachments(t *testing.T) {
	tmpFile := t.TempDir() + "/attach.txt"
	require.NoError(t, os.WriteFile(tmpFile, []byte("x"), 0o644))

	cfg := parseTestConfig(t, []byte(`
[general]
from = "a <a@a.com>"
subject = "test"
attachments = ["`+tmpFile+`"]
[[recipients]]
email = "a@b.com"
first = "Alice"
`))
	assert.Equal(t, []string{tmpFile}, cfg.Attachments)
}

func TestParseWithMissingAttachment(t *testing.T) {
	_, err := Parse([]byte(`
[general]
from = "a <a@a.com>"
subject = "test"
attachments = ["/nonexistent/file.txt"]
[[recipients]]
email = "a@b.com"
first = "Alice"
`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment not found")
}

func TestSampleConfigParses(t *testing.T) {
	cfg, err := Parse([]byte(SampleConfig("0.0.0")))
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.From)
	assert.NotEmpty(t, cfg.Subject)
	assert.NotEmpty(t, cfg.Recipients)
	for _, r := range cfg.Recipients {
		assert.NotEmpty(t, r.Email)
		assert.NotEmpty(t, r.First)
	}
}

func TestSampleTemplateNotEmpty(t *testing.T) {
	tmpl := SampleTemplate()
	assert.NotEmpty(t, tmpl)
	assert.Contains(t, tmpl, "%FN%")
	assert.Contains(t, tmpl, "%LN%")
}
