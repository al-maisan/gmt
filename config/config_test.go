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
	c, err := New(input)
	require.NoError(t, err)
	cfg, err := c.Parse()
	require.NoError(t, err)
	return cfg
}

func TestLoadDefault(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
From=Frodo Baggins <rts@example.com>
subject=Hello %FN%!
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney
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

func TestLoadNoGeneralSection(t *testing.T) {
	c, err := New([]byte(`
[recipients]
jd@example.com=John Doe
`))
	require.NoError(t, err)
	_, err = c.ParseGeneral()
	require.Error(t, err)
	assert.Equal(t, "section not found", err.Error())
}

func TestLoadNoRecipients(t *testing.T) {
	c, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
subject=Hello %FN%!
`))
	require.NoError(t, err)
	cfg, err := c.ParseGeneral()
	require.NoError(t, err)
	err = c.ParseRecipients(&cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipients")
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"missing subject", "[general]\nFrom=x <x@x.com>", "missing required key 'subject'"},
		{"missing from", "[general]\nsubject=test", "missing required key 'from'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New([]byte(tt.input))
			require.NoError(t, err)
			_, err = c.ParseGeneral()
			require.Error(t, err)
			assert.Equal(t, tt.wantErr, err.Error())
		})
	}
}

func TestLoadSubjectCaseInsensitive(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
From=Frodo Baggins <rts@example.com>
Subject=Hello %FN%!
[recipients]
jd@example.com=John Doe
`))
	assert.Equal(t, "Hello %FN%!", cfg.Subject)
}

func TestLoadFull(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
From=Frodo Baggins <rts@example.com>
Cc=weirdo@nsb.gov, cc@example.com
Reply-To=John Doe <jd@mail.com>
subject=Hello %FN%!
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney
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

func TestParseRecipientsSingleName(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
From=test <t@example.com>
subject=test
[recipients]
madonna@example.com=Madonna
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, "Madonna", cfg.Recipients[0].First)
	assert.Equal(t, "", cfg.Recipients[0].Last)
	assert.Equal(t, "madonna@example.com", cfg.Recipients[0].Email)
}

func TestParseRecipientsMalformedData(t *testing.T) {
	cfg := parseTestConfig(t, []byte(`
[general]
From=test <t@example.com>
subject=test
[recipients]
jd@example.com=John Doe|BADDATA|ORG:-EFF
`))
	require.Len(t, cfg.Recipients, 1)
	assert.Equal(t, map[string]string{"ORG": "EFF"}, cfg.Recipients[0].Data)
	require.Len(t, cfg.Warnings, 1)
	assert.Contains(t, cfg.Warnings[0], "jd@example.com")
	assert.Contains(t, cfg.Warnings[0], "BADDATA")
}

func TestSampleConfigParses(t *testing.T) {
	c, err := New([]byte(SampleConfig("0.0.0")))
	require.NoError(t, err)
	cfg, err := c.Parse()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.From)
	assert.NotEmpty(t, cfg.Subject)
	assert.True(t, len(cfg.Recipients) > 0)
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
