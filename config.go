package main

import (
	"github.com/go-ini/ini"
)

type Name struct {
	First string
	Middle []string
	Last string
}

type Recipient struct {
	Email string
	Name
	Data map[string]string
}

type Attachment struct {
	Email string
	Path string
}

type General struct {
	MailProg string
	AttachmentPath string
	EncryptAttachments bool
	SenderEmail string
	SenderName string
	Cc string
	Recipients []Recipient
	Attachments []Attachment
}

func Make(bs []byte) (result General) {
	cfg, err := ini.Load(bs)
	if err != nil {
		return
	}
	sec, err := cfg.GetSection("general")
	if err != nil {
		return
	}
	// mandatory keys
	key, err := sec.GetKey("mail-prog")
	if err != nil {
		return
	}
	result.MailProg = key.String()
	// optional keys
	key, err = sec.GetKey("attachment-path")
	if err == nil {
		result.AttachmentPath = key.String()
	}
	key, err = sec.GetKey("encrypt-attachments")
	if err == nil {
		result.EncryptAttachments, err = key.Bool()
		if err != nil {
			result.EncryptAttachments = false
		}
	}
	key, err = sec.GetKey("sender-email")
	if err == nil {
		result.SenderEmail = key.String()
	}
	key, err = sec.GetKey("sender-name")
	if err == nil {
		result.SenderName = key.String()
	}
	key, err = sec.GetKey("Cc")
	if err == nil {
		result.Cc = key.String()
	}
	return
}
