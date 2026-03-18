// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019-2023  "Muharem Hrnjadovic" <gmt@lbox.cc>
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

	"github.com/go-ini/ini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sortRecipients(r []Recipient) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].Email > r[j].Email
	})
}

func TestLoadDefault(t *testing.T) {
	cfg, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
#Cc=weirdo@nsb.gov, cc@example.com
#Reply-to=John Doe <jd@mail.com>
subject=Hello %FN%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`))
	require.NoError(t, err)

	assert.Equal(t, "Frodo Baggins <rts@example.com>", cfg.From)
	assert.Empty(t, cfg.Cc)

	expected := []Recipient{
		{
			Email: "jd@example.com",
			First: "John",
			Last:  "Doe Jr.",
			Data:  map[string]string{"ORG": "EFF", "TITLE": "PhD"},
		},
		{
			Email: "mm@gmail.com",
			First: "Mickey",
			Last:  "Mouse",
			Data:  map[string]string{"ORG": "Disney"},
		},
	}
	sortRecipients(expected)
	sortRecipients(cfg.Recipients)
	assert.Equal(t, expected, cfg.Recipients)
}

func TestLoadNoGeneralSection(t *testing.T) {
	_, err := New([]byte(`
[recipients]
jd@example.com=John Doe
`))
	require.Error(t, err)
	assert.Equal(t, "config file must have a [general] section", err.Error())
}

func TestLoadNoRecipients(t *testing.T) {
	_, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
subject=Hello %FN%!
`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipients")
}

func TestLoadNoSubject(t *testing.T) {
	_, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
`))
	require.Error(t, err)
	assert.Equal(t, "'subject' not configured", err.Error())
}

func TestLoadNoFrom(t *testing.T) {
	_, err := New([]byte(`
[general]
subject=Hello %FN%!
`))
	require.Error(t, err)
	assert.Equal(t, "'from' not configured", err.Error())
}

func TestLoadSubjectCaseInsensitive(t *testing.T) {
	cfg, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
Subject=Hello %FN%!
[recipients]
jd@example.com=John Doe
`))
	require.NoError(t, err)
	assert.Equal(t, "Hello %FN%!", cfg.Subject)
}

func TestLoadFull(t *testing.T) {
	cfg, err := New([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
Cc=weirdo@nsb.gov, cc@example.com
Reply-To=John Doe <jd@mail.com>
subject=Hello %FN%!
#attachments=/home/user/atmt1.ics, ../Documents/doc2.txt
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`))
	require.NoError(t, err)

	assert.Equal(t, "Frodo Baggins <rts@example.com>", cfg.From)
	assert.Equal(t, "John Doe <jd@mail.com>", cfg.ReplyTo)
	assert.Equal(t, "Hello %FN%!", cfg.Subject)
	assert.Equal(t, []string{"weirdo@nsb.gov", "cc@example.com"}, cfg.Cc)

	expected := []Recipient{
		{
			Email: "jd@example.com",
			First: "John",
			Last:  "Doe Jr.",
			Data:  map[string]string{"ORG": "EFF", "TITLE": "PhD"},
		},
		{
			Email: "mm@gmail.com",
			First: "Mickey",
			Last:  "Mouse",
			Data:  map[string]string{"ORG": "Disney"},
		},
	}
	sortRecipients(expected)
	sortRecipients(cfg.Recipients)
	assert.Equal(t, expected, cfg.Recipients)
}

func TestParseRecipients(t *testing.T) {
	cfg, err := ini.InsensitiveLoad([]byte(`
[general]
From=Frodo Baggins <rts@example.com>
Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
`))
	require.NoError(t, err)

	recipients, err := cfg.GetSection("recipients")
	require.NoError(t, err)

	actual := parseRecipients(recipients)
	require.NotNil(t, actual)

	expected := []Recipient{
		{
			Email: "jd@example.com",
			First: "John",
			Last:  "Doe Jr.",
			Data:  map[string]string{"ORG": "EFF", "TITLE": "PhD"},
		},
		{
			Email: "mm@gmail.com",
			First: "Mickey",
			Last:  "Mouse",
			Data:  map[string]string{"ORG": "Disney"},
		},
	}
	sortRecipients(expected)
	sortRecipients(actual)
	assert.Equal(t, expected, actual)
}

func TestParseRecipientsSingleName(t *testing.T) {
	cfg, err := ini.InsensitiveLoad([]byte(`
[recipients]
madonna@example.com=Madonna
`))
	require.NoError(t, err)

	recipients, err := cfg.GetSection("recipients")
	require.NoError(t, err)

	actual := parseRecipients(recipients)
	require.Len(t, actual, 1)
	assert.Equal(t, "Madonna", actual[0].First)
	assert.Equal(t, "", actual[0].Last)
	assert.Equal(t, "madonna@example.com", actual[0].Email)
}

func TestParseRecipientsMalformedData(t *testing.T) {
	cfg, err := ini.InsensitiveLoad([]byte(`
[recipients]
jd@example.com=John Doe|BADDATA|ORG:-EFF
`))
	require.NoError(t, err)

	recipients, err := cfg.GetSection("recipients")
	require.NoError(t, err)

	actual := parseRecipients(recipients)
	require.Len(t, actual, 1)
	assert.Equal(t, map[string]string{"ORG": "EFF"}, actual[0].Data)
}

func TestSampleConfigParses(t *testing.T) {
	cfg, err := New([]byte(SampleConfig("0.0.0")))
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.From)
	assert.NotEmpty(t, cfg.Subject)
	assert.True(t, len(cfg.Recipients) > 0, "sample config should have at least one recipient")
	for _, r := range cfg.Recipients {
		assert.NotEmpty(t, r.Email, "recipient email must not be empty")
		assert.NotEmpty(t, r.First, "recipient first name must not be empty")
	}
}

func TestSampleTemplateNotEmpty(t *testing.T) {
	tmpl := SampleTemplate("0.0.0")
	assert.NotEmpty(t, tmpl)
	assert.Contains(t, tmpl, "%FN%")
	assert.Contains(t, tmpl, "%LN%")
}
