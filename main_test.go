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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRecipientData(t *testing.T) {

	// assert equality
	name, addr := parseRecipientData(`"abc ähm" <abc@example.com>`)
	assert.Equal(t, "abc ähm", name, "utf-8 name matches")
	assert.Equal(t, "abc@example.com", addr, "email address matches")
}
