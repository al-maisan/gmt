// gmt-mail sends emails in bulk based on a template and a config file.
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
	"runtime/debug"
	"testing"

	"github.com/al-maisan/gmt-mail/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	assert.Contains(t, version(), "dev")
}

func TestFromBuildInfo(t *testing.T) {
	tests := []struct {
		name    string
		bi      debug.BuildInfo
		v, c, d string
	}{
		{"go install @version", debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, "1.2.3", "unknown", "unknown"},
		{"go build in a clone", debug.BuildInfo{Main: debug.Module{Version: "v0.9.1-0.20261031114506-bc86da30ccfc"}, Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "bc86da30ccfc1234567890abcdef"}, {Key: "vcs.time", Value: "2026-10-31T11:45:06Z"},
		}}, "dev", "bc86da3", "2026-10-31"},
		{"go build in a dirty clone", debug.BuildInfo{Main: debug.Module{Version: "v0.9.1-0.20261031114506-bc86da30ccfc+dirty"}}, "dev", "bc86da3+dirty", "2026-10-31"},
		{"go install @commit, no tags", debug.BuildInfo{Main: debug.Module{Version: "v0.0.0-20260102030405-abcdef123456"}}, "dev", "abcdef1", "2026-01-02"},
		{"go install @commit, after a pre-release", debug.BuildInfo{Main: debug.Module{Version: "v1.0.0-rc.1.0.20260102030405-abcdef123456"}}, "dev", "abcdef1", "2026-01-02"},
		{"tagged clone", debug.BuildInfo{Main: debug.Module{Version: "v0.9.1"}}, "0.9.1", "unknown", "unknown"},
		{"dirty tagged clone", debug.BuildInfo{Main: debug.Module{Version: "v0.9.9+dirty"}, Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0f7c75b0123456789"}, {Key: "vcs.time", Value: "2026-10-31T12:00:00Z"},
		}}, "0.9.9", "0f7c75b+dirty", "2026-10-31"},
		{"no VCS stamp (worktree, -buildvcs=false)", debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "69830cd8a1b2c3d4e5f6"}, {Key: "vcs.time", Value: "2026-10-31T11:00:00Z"},
		}}, "dev", "69830cd", "2026-10-31"},
		{"short revision", debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abc"}}}, "dev", "abc", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, c, d := fromBuildInfo(&tt.bi, "dev", "unknown", "unknown")
			assert.Equal(t, []string{tt.v, tt.c, tt.d}, []string{v, c, d})
		})
	}
}

func TestLoadConfigValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
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
	path := filepath.Join(dir, "bad.toml")
	require.NoError(t, os.WriteFile(path, []byte("[general]\nfrom = \"x\"\n"), 0o644))

	_, err := loadConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestPrepMailsValid(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "template.eml")
	require.NoError(t, os.WriteFile(tmplPath, []byte(config.SampleTemplate()), 0o644))

	cfg, err := config.Parse([]byte(config.SampleConfig("0.0.0")))
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
