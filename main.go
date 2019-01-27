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
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/al-maisan/gmt/config"
	"github.com/al-maisan/gmt/email"
)

func help() {
	fmt.Fprintf(flag.CommandLine.Output(), "\n%s, version %s\nThis tool sends emails in bulk based on a template and a config file\n\n", filepath.Base(os.Args[0]), Version())
	flag.PrintDefaults()
}

func Version() string { return "0.1.9" }

func main() {

	flag.Usage = help
	configPath := flag.String("config-path", "", "path to the config file")
	doDryRun := flag.Bool("dry-run", false, "show what would be done but execute no action")
	templatePath := flag.String("template-path", "", "path to the template file")
	doSampleConfig := flag.Bool("sample-config", false, "output sample configuration to stdout")
	doSampleTemplate := flag.Bool("sample-template", false, "output sample template to stdout")

	flag.Parse()

	if *doSampleConfig {
		fmt.Println(config.SampleConfig(Version()))
		os.Exit(0)
	}

	if *doSampleTemplate {
		fmt.Println(config.SampleTemplate(Version()))
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "*** Error: please specify config file!")
		flag.Usage()
		os.Exit(1)
	}
	if *templatePath == "" {
		fmt.Fprintln(os.Stderr, "*** Error: please specify template file!")
		flag.Usage()
		os.Exit(2)
	}

	// read the config file
	bytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	// parse the config file
	var cfg config.Data
	cfg, err = config.New(bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in config file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	// read the template file
	bytes, err = ioutil.ReadFile(*templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read template file!")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	// prepare the emails, substitute vars in subject & body
	mails := email.PrepMails(cfg, string(bytes))

	// is this a dry run? print what would be done if so and exit
	if *doDryRun == true {
		for addr, data := range mails {
			cmdline := []string{}
			var body string

			if cfg.MailProg == "sendmail" {
				body, cmdline = prepSendmailBody(addr, data, cmdline[1:])
			} else {
				body = data.Body
			}
			fmt.Fprintf(os.Stdout, "--\n%s\n%s\n", cmdline, body)
		}
		os.Exit(0)
	}

	send(mails, cfg)
}

func send(mails map[string]email.Data, cfg config.Data) {
	ch := make(chan string)
	for addr, data := range mails {
		go sendEmail(addr, data, cfg, ch)
	}
	fmt.Println("\nSending emails now..")
	for i := len(mails); i > 0; i-- {
		msg := <-ch
		fmt.Println(msg)
	}
	return
}

func prepSendmailBody(addr string, data email.Data, cmdline []string) (body string, new_cmdline []string) {
	lines := []string{fmt.Sprintf("To: %s", addr)}
	if data.Subject != "" {
		lines = append(lines, fmt.Sprintf("Subject: %s", data.Subject))
	}
	for i := 0; i < len(cmdline); i++ {
		lines = append(lines, fmt.Sprintf("%s %s", cmdline[i], cmdline[i+1]))
		i++
	}
	lines = append(lines, fmt.Sprintf("X-Mailer: gmt, version %s, https://301.mx/gmt", Version()))

	header := strings.Join(lines, "\n")
	body = strings.Join([]string{header, data.Body}, "\n\n")
	new_cmdline = []string{"sendmail", "-t"}

	return
}

func sendEmail(addr string, data email.Data, cfg config.Data, ch chan string) {
	// prepare the command line args for the mail user agent (MUA)
	cmdline := []string{}

	var body string
	var cmd2args []string

	if cfg.MailProg == "sendmail" {
		body, cmdline = prepSendmailBody(addr, data, cmdline[1:])
		cmd2args = cmdline[1:]
	} else {
		body = data.Body
		cmd2args = append(cmdline[1:], "-s", data.Subject, addr)
	}

	file, err := tempFile([]byte(body))
	if err != nil {
		ch <- fmt.Sprintf("!! Error sending to %s (%s)", addr, err.Error())
		return
	}
	defer os.Remove(file)

	cmd1 := exec.Command("cat", file)
	cmd2 := exec.Command(cmdline[0], cmd2args...)
	if _, err = pipeCmds(cmd1, cmd2); err != nil {
		ch <- fmt.Sprintf("!! Error sending to %s (%s)", addr, err.Error())
		return
	} else {
		ch <- fmt.Sprintf("-> %s", addr)
	}
}

func tempFile(content []byte) (name string, err error) {
	var tmpfile *os.File
	tmpfile, err = ioutil.TempFile("", "gmt")
	if err != nil {
		return
	}
	if _, err = tmpfile.Write(content); err != nil {
		return
	}
	if err = tmpfile.Close(); err != nil {
		return
	}
	name = tmpfile.Name()
	return
}

func pipeCmds(cmd1, cmd2 *exec.Cmd) (result string, err error) {
	reader, writer := io.Pipe()

	// push first command output to writer
	cmd1.Stdout = writer

	// read from first command output
	cmd2.Stdin = reader

	// prepare a buffer to capture the output
	// after second command finished executing
	var buf bytes.Buffer
	cmd2.Stdout = &buf

	if err = cmd1.Start(); err != nil {
		err = errors.New(fmt.Sprintf("cmd1 start failure (%s -- %s)", cmd1.Args, err.Error()))
		return
	}
	if err = cmd2.Start(); err != nil {
		err = errors.New(fmt.Sprintf("cmd2 start failure (%s -- %s)", cmd2.Args, err.Error()))
		return
	}
	if err = cmd1.Wait(); err != nil {
		err = errors.New(fmt.Sprintf("cmd1 wait failure (%s -- %s)", cmd1.Args, err.Error()))
		return
	}
	writer.Close()
	if err = cmd2.Wait(); err != nil {
		err = errors.New(fmt.Sprintf("cmd2 wait failure (%s -- %s)", cmd2.Args, err.Error()))
		return
	}

	result = buf.String()
	return
}
