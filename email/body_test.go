package email

import (
	"testing"

	"github.com/al-maisan/gmt/config"
	"github.com/stretchr/testify/assert"
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
