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

func TestSubstVarsBasic(t *testing.T) {
	r := config.Recipient{
		Email: "jd@example.com",
		First: "John",
		Last:  "Doe",
		Data:  map[string]string{"ORG": "EFF"},
	}
	result := substVars(r, "Hello %FN% %LN%, you work at %ORG%. Email: %EA%")
	assert.Equal(t, "Hello John Doe, you work at EFF. Email: jd@example.com", result)
}

func TestSubstVarsNoCustomData(t *testing.T) {
	r := config.Recipient{
		Email: "a@b.com",
		First: "Alice",
		Last:  "Bob",
		Data:  map[string]string{},
	}
	result := substVars(r, "%FN% %LN%")
	assert.Equal(t, "Alice Bob", result)
}

func TestSubstVarsUnknownPlaceholder(t *testing.T) {
	r := config.Recipient{
		Email: "a@b.com",
		First: "Alice",
		Last:  "Bob",
		Data:  map[string]string{},
	}
	result := substVars(r, "Hello %FN%, your title is %TITLE%")
	assert.Equal(t, "Hello Alice, your title is %TITLE%", result)
}

func TestPrepMailsBasic(t *testing.T) {
	cfg := config.Data{
		Subject: "Hi %FN%!",
		Recipients: []config.Recipient{
			{Email: "jd@example.com", First: "John", Last: "Doe", Data: map[string]string{}},
		},
	}
	mails := PrepMails(cfg, "Hello %FN% %LN%")
	assert.Len(t, mails, 1)
	assert.Equal(t, "John Doe", mails[0].Name)
	assert.Equal(t, "jd@example.com", mails[0].Address)
}

func TestPrepMailsSingleName(t *testing.T) {
	cfg := config.Data{
		Subject: "Hi %FN%!",
		Recipients: []config.Recipient{
			{Email: "m@example.com", First: "Madonna", Last: "", Data: map[string]string{}},
		},
	}
	mails := PrepMails(cfg, "Hello %FN%")
	assert.Len(t, mails, 1)
	assert.Equal(t, "Madonna", mails[0].Name)
	assert.Equal(t, "m@example.com", mails[0].Address)
	assert.Equal(t, "Hi Madonna!", mails[0].Subject)
	assert.Equal(t, "Hello Madonna", mails[0].Body)
}

func TestPrepMailsGlobalCcCopied(t *testing.T) {
	cfg := config.Data{
		Subject: "Test",
		Cc:      []string{"cc1@example.com", "cc2@example.com"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{}},
		},
	}
	mails := PrepMails(cfg, "body")
	assert.Equal(t, []string{"cc1@example.com", "cc2@example.com"}, mails[0].Cc)
}

func TestPrepMailsPerRecipientCcReplace(t *testing.T) {
	cfg := config.Data{
		Subject: "Test",
		Cc:      []string{"global@cc.com"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"CC": "override@cc.com",
			}},
		},
	}
	mails := PrepMails(cfg, "body")
	assert.Equal(t, []string{"override@cc.com"}, mails[0].Cc)
}

func TestPrepMailsPerRecipientCcAppend(t *testing.T) {
	cfg := config.Data{
		Subject: "Test",
		Cc:      []string{"global@cc.com"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"CC": "+extra@cc.com",
			}},
		},
	}
	mails := PrepMails(cfg, "body")
	assert.Equal(t, []string{"global@cc.com", "extra@cc.com"}, mails[0].Cc)
}

func TestPrepMailsPerRecipientAttachmentReplace(t *testing.T) {
	cfg := config.Data{
		Subject:     "Test",
		Attachments: []string{"global.txt"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"AS": "local.txt",
			}},
		},
	}
	mails := PrepMails(cfg, "body")
	assert.Equal(t, []string{"local.txt"}, mails[0].Attachments)
}

func TestPrepMailsPerRecipientAttachmentAppend(t *testing.T) {
	cfg := config.Data{
		Subject:     "Test",
		Attachments: []string{"global.txt"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"AS": "+extra.txt",
			}},
		},
	}
	mails := PrepMails(cfg, "body")
	assert.Equal(t, []string{"global.txt", "extra.txt"}, mails[0].Attachments)
}

func TestPrepMailsCcNotUsedAsTemplateVar(t *testing.T) {
	cfg := config.Data{
		Subject: "Test",
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"CC": "someone@cc.com",
			}},
		},
	}
	mails := PrepMails(cfg, "cc is %CC%")
	// CC should be extracted, not substituted as a template var
	assert.Equal(t, "cc is %CC%", mails[0].Body)
}

func TestPrepMailsDoesNotMutateConfig(t *testing.T) {
	cfg := config.Data{
		Subject: "Test",
		Cc:      []string{"global@cc.com"},
		Recipients: []config.Recipient{
			{Email: "a@b.com", First: "A", Last: "B", Data: map[string]string{
				"CC": "override@cc.com", "ORG": "EFF",
			}},
		},
	}
	PrepMails(cfg, "body")
	// original Data map must still have CC and ORG
	assert.Equal(t, "override@cc.com", cfg.Recipients[0].Data["CC"])
	assert.Equal(t, "EFF", cfg.Recipients[0].Data["ORG"])
}

func TestResolveOverrideNoOverride(t *testing.T) {
	result := resolveOverride([]string{"a", "b"}, map[string]string{}, "CC")
	assert.Equal(t, []string{"a", "b"}, result)
}

func TestResolveOverrideReplace(t *testing.T) {
	result := resolveOverride([]string{"a", "b"}, map[string]string{"CC": "x,y"}, "CC")
	assert.Equal(t, []string{"x", "y"}, result)
}

func TestResolveOverrideAppend(t *testing.T) {
	result := resolveOverride([]string{"a"}, map[string]string{"CC": "+b,c"}, "CC")
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestResolveOverrideEmptyGlobal(t *testing.T) {
	result := resolveOverride(nil, map[string]string{"CC": "x"}, "CC")
	assert.Equal(t, []string{"x"}, result)
}

func TestSampleConfigAndTemplateIntegration(t *testing.T) {
	c, err := config.New([]byte(config.SampleConfig("0.0.0")))
	require.NoError(t, err)

	cfg, err := c.ParseGeneral()
	require.NoError(t, err)

	cfg.Recipients, err = c.ParseRecipients()
	require.NoError(t, err)

	tmpl := config.SampleTemplate()
	mails := PrepMails(cfg, tmpl)

	assert.True(t, len(mails) > 0, "should produce at least one mail")
	for _, m := range mails {
		assert.NotEmpty(t, m.Name, "mail name must not be empty")
		assert.NotEmpty(t, m.Address, "mail address must not be empty")
		assert.NotEmpty(t, m.Subject, "mail subject must not be empty")
		assert.NotEmpty(t, m.Body, "mail body must not be empty")
		assert.NotContains(t, m.Subject, "%FN%", "subject should have FN substituted")
		assert.NotContains(t, m.Body, "%FN%", "body should have FN substituted")
		assert.NotContains(t, m.Body, "%LN%", "body should have LN substituted")
		assert.NotContains(t, m.Body, "%ORG%", "body should have ORG substituted")
	}
}
