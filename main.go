package main

import (
	"flag"
	"fmt"
)

func main() {
	configPath := flag.String("config-path", "", "path to the config file")
	doDryRun := flag.Bool("dry-run", false, "show what would be done but execute no action")
	subject := flag.String("subject", "", "email subject")
	templatePath := flag.String("template-path", "", "path to the template file")
	doSampleConfig := flag.Bool("sample-config", false, "output sample configuration to stdout")
	doSampleTemplate := flag.Bool("sample-template", false, "output sample template to stdout")

	flag.Parse()

	if *doSampleConfig {
		fmt.Println(sampleConfig())
		return
	}

	if *doSampleTemplate {
		fmt.Println(sampleTemplate())
		return
	}

	fmt.Println("configPath: ", *configPath)
	fmt.Println("doDryRun: ", *doDryRun)
	fmt.Println("subject: ", *subject)
	fmt.Println("templatePath: ", *templatePath)
	fmt.Println("doSampleConfig: ", *doSampleConfig)
	fmt.Println("doSampleTemplate: ", *doSampleTemplate)
}

func sampleConfig() string {
	return `# anything that follows a hash is a comment
# email address is to the left of the '=' sign, first word after is
# the first name, the rest is the surname
[general]
mail-prog=gnu-mail # arch linux, 'mail' on ubuntu, 'mailx' on Fedora
#attachment-path=/tmp
#encrypt-attachments=true
sender-email=rts@example.com
sender-name=Frodo Baggins
#Cc=weirdo@nsb.gov, cc@example.com
[recipients]
jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
mm@gmail.com=Mickey Mouse|ORG:-Disney   # trailing comment!!
[attachments]
jd@example.com=01.pdf
mm@gmail.com=02.pdf`
}

func sampleTemplate() string {
	return `FN / LN / EA = first name / last name / email address

Hello %FN% // %LN%, how are things going at %ORG%?
this is your email * 2: %EA%%EA%.`
}
