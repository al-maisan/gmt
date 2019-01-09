# Go mailing tool (gmt)

## What is it?
`gmt` is a simple utility that allows the automated sending of emails using a configuration file and a template for the email body.
It was written in [Go](https://golang.org/) and used mainly on linux systems like [Arch Linux](https://www.archlinux.org/), [Fedora](https://getfedora.org/) and [Ubuntu](http://www.ubuntu.com/).

## Prerequisites
Running `gmt` requires a running mail transfer agent (MTA) e.g. [postfix](http://www.postfix.org/) and the [GNU mailutils](https://www.gnu.org/software/mailutils/mailutils.html) software (Arch Linux: [`pacman -S mailutils`](https://www.archlinux.org/packages/?sort=&q=mailutils&maintainer=&flagged=), Ubuntu: [`apt-get install mailutils`](http://packages.ubuntu.com/search?keywords=mailutils)). For Fedora etc. you are all set since `mailx` is installed by default.

## Using `gmt`
The easiest way to use the tool is to generate a sample configuration (`-sample-config`) and a template file (`-sample-template`) and take it from there.

    $ ./gmt -h

     gmt sends emails in bulk based on a template and a config file

       -config-path string
             path to the config file
       -dry-run
             show what would be done but execute no action
       -sample-config
             output sample configuration to stdout
       -sample-template
             output sample template to stdout
       -template-path string
             path to the template file
