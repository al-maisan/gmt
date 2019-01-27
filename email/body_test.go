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

package email

import (
	"testing"

	"github.com/al-maisan/gmt/config"
	. "github.com/smartystreets/goconvey/convey"
)

func GetConfig(input string) config.Data {
	cfg, _ := config.New([]byte(input))
	return cfg
}

func TestPrepBodies(t *testing.T) {
	Convey("generate email bodies for default config", t, func() {
		version := "0.2.0"
		cfg := GetConfig(config.SampleConfig(version))
		template := config.SampleTemplate(version)
		bodies := PrepMails(cfg, template)

		expected := `FN / LN / EA = first name / last name / email address

Hello John // Doe Jr., how are things going at EFF?
this is your email: jd@example.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		So(bodies["jd@example.com"].Body, ShouldEqual, expected)
		So(bodies["jd@example.com"].Subject, ShouldEqual, "Hello John!")

		expected = `FN / LN / EA = first name / last name / email address

Hello Mickey // Mouse, how are things going at Disney?
this is your email: mm@gmail.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		So(bodies["mm@gmail.com"].Body, ShouldEqual, expected)
		So(bodies["mm@gmail.com"].Subject, ShouldEqual, "Hello Mickey!")

		expected = `FN / LN / EA = first name / last name / email address

Hello Daisy // Lila, how are things going at NASA?
this is your email: daisy@example.com :)


Sent with gmt version 0.2.0, see https://301.mx/gmt for details.`
		So(bodies["daisy@example.com"].Body, ShouldEqual, expected)
		So(bodies["daisy@example.com"].Subject, ShouldEqual, "Hello Daisy!")
	})
}
