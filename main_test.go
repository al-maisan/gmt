// gmt sends emails in bulk based on a template and a config file.
// Copyright (C) 2019  "Muharem Hrnjadovic" <gmt@lbox.cc>
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
	"os/exec"
	"strings"
	"testing"

	"github.com/al-maisan/gmt/config"
	. "github.com/smartystreets/goconvey/convey"
)

func GetConfig(input string) config.Data {
	cfg, _ := config.New([]byte(input))
	return cfg
}

func TestPipeCmds(t *testing.T) {
	Convey("Test pipeCmds() with  cat / wc", t, func() {
		fname, err := tempFile([]byte("line1\nline2\nline3\n"))
		So(err, ShouldBeNil)

		defer os.Remove(fname)
		cmd1 := exec.Command("cat", fname)
		cmd2 := exec.Command("wc", "-l")
		result, err := pipeCmds(cmd1, cmd2)

		So(err, ShouldBeNil)
		So(strings.TrimSpace(result), ShouldEqual, "3")
	})
}

func TestPipeCmdsWithCmd1Fail(t *testing.T) {
	Convey("Test pipeCmds() with  cmd1 failure", t, func() {
		fname, err := tempFile([]byte("line1\nline2\nline3\n"))
		So(err, ShouldBeNil)

		defer os.Remove(fname)
		cmd1 := exec.Command("ls", "-cdjkgfrgf")
		cmd2 := exec.Command("wc", "-l")
		_, err = pipeCmds(cmd1, cmd2)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "cmd1 wait failure ([ls -cdjkgfrgf] -- exit status 2)")
	})
}

func TestPipeCmdsWithCmd2Fail(t *testing.T) {
	Convey("Test pipeCmds() with  cmd2 failure", t, func() {
		fname, err := tempFile([]byte("line1\nline2\nline3\n"))
		So(err, ShouldBeNil)

		defer os.Remove(fname)
		cmd1 := exec.Command("ls", "-l")
		cmd2 := exec.Command("wc", "-dksvgdk")
		_, err = pipeCmds(cmd1, cmd2)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "cmd2 wait failure ([wc -dksvgdk] -- exit status 1)")
	})
}
