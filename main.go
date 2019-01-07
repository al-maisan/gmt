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
	"os"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
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

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Please specify config file!")
		os.Exit(1)
	}
	if *templatePath == "" {
		fmt.Fprintln(os.Stderr, "Please specify template file!")
		os.Exit(2)
	}

	bytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	var cfg config.Data

	cfg, err = config.New(bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	args := email.PrepMUAArgs(cfg)
	fmt.Println(args)

	bytes, err = ioutil.ReadFile(*templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read template file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	emails := email.PrepMails(cfg, string(bytes))
	if *doDryRun == true {
		for ea, mail := range emails {
			fmt.Fprintf(os.Stdout, "--\nTo: %s\nSubject: %s\n%s\n", ea, mail.Subject, mail.Body)
		}
		os.Exit(0)
	}
}
