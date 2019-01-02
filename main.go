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
	"io/ioutil"
	"flag"
	"fmt"

	"github.com/al-maisan/gmt/config"
)

func main() {
	configPath := flag.String("config-path", "", "path to the config file")
	doDryRun := flag.Bool("dry-run", false, "show what would be done but execute no action")
	templatePath := flag.String("template-path", "", "path to the template file")
	doSampleConfig := flag.Bool("sample-config", false, "output sample configuration to stdout")
	doSampleTemplate := flag.Bool("sample-template", false, "output sample template to stdout")

	flag.Parse()

	if *doSampleConfig {
		fmt.Println(config.SampleConfig())
		return
	}

	if *doSampleTemplate {
		fmt.Println(config.SampleTemplate())
		return
	}

	fmt.Println("configPath: ", *configPath)
	fmt.Println("doDryRun: ", *doDryRun)
	fmt.Println("templatePath: ", *templatePath)
	fmt.Println("doSampleConfig: ", *doSampleConfig)
	fmt.Println("doSampleTemplate: ", *doSampleTemplate)

	if configPath != nil {
		bytes, err := ioutil.ReadFile(*configPath)

		if err == nil {
			cfg, cerr := config.New(bytes)
			if cerr != nil {
				fmt.Println(cerr)
			} else {
				fmt.Printf("%+v\n", cfg)
			}
		}
	}
}
