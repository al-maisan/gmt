# Go mailing tool (gmt)

## What is it?
`gmt` is a simple utility that allows the automated sending of emails using a configuration file and a template for the email body.
It was written in [Go](https://golang.org/) and used mainly on linux systems like [Arch Linux](https://www.archlinux.org/), [Fedora](https://getfedora.org/) and [Ubuntu](http://www.ubuntu.com/).

## Prerequisites
Running `gmt` requires a running mail transfer agent (MTA) e.g. [postfix](http://www.postfix.org/) and the [GNU mailutils](https://www.gnu.org/software/mailutils/mailutils.html) software (Arch Linux: [`pacman -S mailutils`](https://www.archlinux.org/packages/?sort=&q=mailutils&maintainer=&flagged=), Ubuntu: [`apt-get install mailutils`](http://packages.ubuntu.com/search?keywords=mailutils)). For Fedora etc. you are all set since `mailx` is installed by default.

## Using `gmt`
The easiest way to use the tool is to generate a sample configuration (`-sample-config`) and a template file (`-sample-template`) and take it from there.

    $ go build
    $ ./gmt -sample-config > /tmp/sc.ini
    $ ./gmt -sample-template > /tmp/st.eml
    $ ./gmt -dry-run -config-path /tmp/sc.ini  -template-path /tmp/st.eml
    --
    To: jd@example.com
    Subject: Hello John!
    FN / LN / EA = first name / last name / email address

    Hello John // Doe Jr., how are things going at EFF?
    this is your email * 2: jd@example.comjd@example.com.

    --
    To: mm@gmail.com
    Subject: Hello Mickey!
    FN / LN / EA = first name / last name / email address

    Hello Mickey // Mouse, how are things going at Disney?
    this is your email * 2: mm@gmail.commm@gmail.com.


Last but not least use `-h` to see all the options:


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
