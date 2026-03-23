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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/al-maisan/gmt/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	assert.Contains(t, version(), "dev")
}

func TestLoadConfigValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.ini")
	require.NoError(t, os.WriteFile(path, []byte(config.SampleConfig("0.0.0")), 0o644))

	cfg, err := loadConfig(path)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.From)
	assert.NotEmpty(t, cfg.Subject)
	assert.NotEmpty(t, cfg.Recipients)
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := loadConfig("/nonexistent/config.ini")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestLoadConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.ini")
	require.NoError(t, os.WriteFile(path, []byte("[general]\n"), 0o644))

	_, err := loadConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestPrepMailsValid(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "template.eml")
	require.NoError(t, os.WriteFile(tmplPath, []byte(config.SampleTemplate()), 0o644))

	c, err := config.New([]byte(config.SampleConfig("0.0.0")))
	require.NoError(t, err)
	cfg, err := c.Parse()
	require.NoError(t, err)

	msgs, err := prepMails(&cfg, tmplPath)
	require.NoError(t, err)
	assert.NotEmpty(t, msgs)
}

func TestPrepMailsMissingTemplate(t *testing.T) {
	cfg := config.MailConfig{}
	_, err := prepMails(&cfg, "/nonexistent/template.eml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read template")
}
