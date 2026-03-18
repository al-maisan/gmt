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

	"github.com/al-maisan/gmt/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstVars(t *testing.T) {
	tests := []struct {
		name   string
		r      config.Recipient
		text   string
		want   string
	}{
		{
			name: "all placeholders",
			r:    config.Recipient{Email: "jd@example.com", First: "John", Last: "Doe", Data: map[string]string{"ORG": "EFF"}},
			text: "Hello %FN% %LN%, you work at %ORG%. Email: %EA%",
			want: "Hello John Doe, you work at EFF. Email: jd@example.com",
		},
		{
			name: "no custom data",
			r:    config.Recipient{Email: "a@b.com", First: "Alice", Last: "Bob", Data: map[string]string{}},
			text: "%FN% %LN%",
			want: "Alice Bob",
		},
		{
			name: "unknown placeholder left as-is",
			r:    config.Recipient{Email: "a@b.com", First: "Alice", Last: "Bob", Data: map[string]string{}},
			text: "Hello %FN%, your title is %TITLE%",
			want: "Hello Alice, your title is %TITLE%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, substVars(tt.r, tt.text))
		})
	}
}

func TestPrepMailsBasic(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Hi %FN%!",
		Recipients: []config.Recipient{
			{Email: "jd@example.com", First: "John", Last: "Doe", Data: map[string]string{}},
		},
	}
	mails := PrepMails(&cfg, "Hello %FN% %LN%")
	assert.Len(t, mails, 1)
	assert.Empty(t, cfg.Warnings)
	assert.Equal(t, "John Doe", mails[0].Name)
	assert.Equal(t, "jd@example.com", mails[0].Address)
	assert.Equal(t, "Hi John!", mails[0].Subject)
	assert.Equal(t, "Hello John Doe", mails[0].Body)
}

func TestPrepMailsSingleName(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Hi %FN%!",
		Recipients: []config.Recipient{
			{Email: "m@example.com", First: "Madonna", Last: "", Data: map[string]string{}},
		},
	}
	mails := PrepMails(&cfg, "Hello %FN%")
	assert.Len(t, mails, 1)
	assert.Equal(t, "Madonna", mails[0].Name)
}

func TestPrepMailsPerRecipientOverrides(t *testing.T) {
	tests := []struct {
		name            string
		globalCc        []string
		globalAttach    []string
		data            map[string]string
		wantCc          []string
		wantAttachments []string
	}{
		{
			name:     "cc replace",
			globalCc: []string{"global@cc.com"},
			data:     map[string]string{"CC": "override@cc.com"},
			wantCc:   []string{"override@cc.com"},
		},
		{
			name:     "cc append",
			globalCc: []string{"global@cc.com"},
			data:     map[string]string{"CC": "+extra@cc.com"},
			wantCc:   []string{"global@cc.com", "extra@cc.com"},
		},
		{
			name:     "cc global copied",
			globalCc: []string{"cc1@example.com", "cc2@example.com"},
			data:     map[string]string{},
			wantCc:   []string{"cc1@example.com", "cc2@example.com"},
		},
		{
			name:            "attachment replace",
			globalAttach:    []string{"global.txt"},
			data:            map[string]string{"AS": "local.txt"},
			wantAttachments: []string{"local.txt"},
		},
		{
			name:            "attachment append",
			globalAttach:    []string{"global.txt"},
			data:            map[string]string{"AS": "+extra.txt"},
			wantAttachments: []string{"global.txt", "extra.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.MailConfig{
				Subject:     "Test",
				Cc:          tt.globalCc,
				Attachments: tt.globalAttach,
				Recipients: []config.Recipient{
					{Email: "a@b.com", First: "A", Last: "B", Data: tt.data},
				},
			}
			mails := PrepMails(&cfg, "body")
			require.Len(t, mails, 1)
			if tt.wantCc != nil {
				assert.Equal(t, tt.wantCc, mails[0].Cc)
			}
			if tt.wantAttachments != nil {
				assert.Equal(t, tt.wantAttachments, mails[0].Attachments)
			}
		})
	}
}

func TestPrepMailsCcNotUsedAsTemplateVar(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Test",
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{"CC": "someone@cc.com"}},
		},
	}
	mails := PrepMails(&cfg, "cc is %CC%")
	assert.Equal(t, "cc is %CC%", mails[0].Body)
}

func TestPrepMailsDoesNotMutateConfig(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Test",
		Cc:      []string{"global@cc.com"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{"CC": "override@cc.com", "ORG": "EFF"}},
		},
	}
	PrepMails(&cfg, "body")
	assert.Equal(t, "override@cc.com", cfg.Recipients[0].Data["CC"])
	assert.Equal(t, "EFF", cfg.Recipients[0].Data["ORG"])
}

func TestPrepMailsUnresolvedPlaceholders(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Hi %FN% from %DEPT%!",
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "Alice", Last: "Bob", Data: map[string]string{}},
		},
	}
	mails := PrepMails(&cfg, "Hello %FN%, your role is %ROLE%")
	assert.Len(t, mails, 1)
	require.Len(t, cfg.Warnings, 2)
	assert.Contains(t, cfg.Warnings[0], "%DEPT%")
	assert.Contains(t, cfg.Warnings[1], "%ROLE%")
}

func TestPrepMailsNoWarningWhenAllResolved(t *testing.T) {
	cfg := config.MailConfig{
		Subject: "Hi %FN%!",
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "Alice", Last: "Bob", Data: map[string]string{"ORG": "EFF"}},
		},
	}
	PrepMails(&cfg, "Hello %FN% from %ORG%")
	assert.Empty(t, cfg.Warnings)
}

func TestResolveOverride(t *testing.T) {
	tests := []struct {
		name   string
		global []string
		data   map[string]string
		want   []string
	}{
		{"no override", []string{"a", "b"}, map[string]string{}, []string{"a", "b"}},
		{"replace", []string{"a", "b"}, map[string]string{"CC": "x,y"}, []string{"x", "y"}},
		{"append", []string{"a"}, map[string]string{"CC": "+b,c"}, []string{"a", "b", "c"}},
		{"empty global", nil, map[string]string{"CC": "x"}, []string{"x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveOverride(tt.global, tt.data, "CC"))
		})
	}
}

func TestSampleConfigAndTemplateIntegration(t *testing.T) {
	c, err := config.New([]byte(config.SampleConfig("0.0.0")))
	require.NoError(t, err)

	cfg, err := c.Parse()
	require.NoError(t, err)

	mails := PrepMails(&cfg, config.SampleTemplate())

	assert.Empty(t, cfg.Warnings, "sample config+template should produce no warnings")
	require.True(t, len(mails) > 0, "should produce at least one mail")
	for _, m := range mails {
		assert.NotEmpty(t, m.Name)
		assert.NotEmpty(t, m.Address)
		assert.NotEmpty(t, m.Subject)
		assert.NotEmpty(t, m.Body)
	}
}
