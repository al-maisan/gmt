# Go mailing tool (gmt-mail)

`gmt-mail` is a simple utility that sends personalized emails in bulk using a configuration file and a template for the email body. It connects directly via SMTP with mandatory TLS.

## Installation

### Ubuntu (PPA)

    $ sudo add-apt-repository ppa:al-maisan/gmt-mail
    $ sudo apt update
    $ sudo apt install gmt-mail

### Fedora (COPR)

    $ sudo dnf copr enable al-maisan/gmt-mail
    $ sudo dnf install gmt-mail

### From source

Requires [Go](https://go.dev/) 1.24 or later.

    $ git clone https://github.com/al-maisan/gmt.git
    $ cd gmt
    $ make build
    $ ./gmt-mail -h

The Makefile embeds the version, git commit hash, and build date into the binary. Targets: `all`, `build`, `test`, `vet`, `lint`, `fmt`, `install`, `clean`, `srpm`.

## Quick start

Generate sample files, configure SMTP credentials, preview, then send:

    $ ./gmt-mail -sample-config > config.toml
    $ ./gmt-mail -sample-template > template.eml
    $ cp .env.example .env    # edit .env with your SMTP credentials
    $ ./gmt-mail -dry-run -config-path config.toml -template-path template.eml
    $ ./gmt-mail -config-path config.toml -template-path template.eml

## SMTP setup

SMTP credentials are read from environment variables. If a `.env` file exists in the working directory, it is loaded automatically.

Copy `.env.example` to `.env` and fill in:

    SMTP_HOST=smtp.gmail.com
    SMTP_PORT=587
    SENDER_EMAIL=your-email@gmail.com
    SENDER_PASSWORD=your-app-password

`SENDER_EMAIL` / `SENDER_PASSWORD` are used for SMTP authentication. The `From` header in the config file controls what recipients see as the sender -- these can differ if your mail server allows it.

For Gmail you must use an [App Password](https://myaccount.google.com/apppasswords) (requires 2-Step Verification). Other providers (Outlook, Yahoo, etc.) are documented in `.env.example`.

TLS is enforced -- credentials are never sent in plaintext.

## Configuration file

The config file uses [TOML](https://toml.io/) format with a `[general]` section and one or more `[[recipients]]` entries.

### `[general]` section

| Key           | Required | Description                                  |
|---------------|----------|----------------------------------------------|
| `from`        | yes      | Sender name and address for the email header  |
| `subject`     | yes      | Email subject line (supports template vars)   |
| `cc`          | no       | List of CC addresses                          |
| `reply_to`    | no       | Reply-To address                              |
| `attachments` | no       | List of file paths to attach                  |

### `[[recipients]]` entries

Each recipient is a TOML table with these fields:

| Key                  | Required | Description                                  |
|----------------------|----------|----------------------------------------------|
| `email`              | yes      | Recipient email address                       |
| `first`              | yes      | First name                                    |
| `last`               | no       | Last name                                     |
| `data`               | no       | Custom key-value pairs for template variables  |
| `cc`                 | no       | Replace global Cc for this recipient           |
| `cc_extra`           | no       | Append to global Cc for this recipient         |
| `attachments`        | no       | Replace global attachments for this recipient  |
| `attachments_extra`  | no       | Append to global attachments for this recipient|

Example:

```toml
[general]
from = '"Frodo Baggins" <rts@example.com>'
subject = "Hello %FN%!"
cc = ["weirdo@nsb.gov", "cc@example.com"]

[[recipients]]
email = "jd@example.com"
first = "John"
last = "Doe Jr."
data = { ORG = "EFF", TITLE = "PhD" }
cc = ["override@cc.com"]

[[recipients]]
email = "mm@gmail.com"
first = "Mickey"
last = "Mouse"
data = { ORG = "Disney" }
```

## Template variables

Templates support these placeholders (in both subject and body). Custom keys are matched in **uppercase** -- use `%ORG%` not `%org%`.

| Placeholder | Value                                    |
|-------------|------------------------------------------|
| `%FN%`      | First name                               |
| `%LN%`      | Last name                                |
| `%EA%`      | Email address                            |
| `%KEY%`     | Any custom key from recipient data       |

Example template:

    Dear %FN% %LN%,

    How are things going at %ORG%?

    Best regards

## Dry run

Use `-dry-run` to preview all emails without sending. The output includes Cc and attachment information when present:

    $ ./gmt-mail -dry-run -config-path config.toml -template-path template.eml
    --
    "John Doe Jr." <jd@example.com>
    Cc: bl@kf.io, info@ex.org
    Subject: Hello John!
    Dear John Doe Jr.,

    How are things going at EFF?

    Best regards
    --
    "Mickey Mouse" <mm@gmail.com>
    Subject: Hello Mickey!
    Dear Mickey Mouse,

    How are things going at Disney?

    Best regards

## Exit codes

| Code | Meaning                          |
|------|----------------------------------|
| 0    | Success                          |
| 1    | Usage error (missing flags)      |
| 2    | Config or template file error    |
| 3    | SMTP connection error            |
| 4    | One or more emails failed to send|

## CLI reference

    $ ./gmt-mail -h

    gmt-mail, version 0.2.1-a1b2c3d (2026-03-18)
    This tool sends emails in bulk based on a template and a config file

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
      -version
            print version and exit

## Releasing a new version

Update `VERSION` in the Makefile, commit, then publish everywhere:

    $ make publish

This tags the release, creates a GitHub release, submits to Fedora COPR, and uploads to Ubuntu PPA.

Individual targets are also available:

| Target    | Description                                    |
|-----------|------------------------------------------------|
| `tag`     | Create and push a signed git tag               |
| `release` | Tag + create GitHub release                    |
| `srpm`    | Build a source RPM                             |
| `copr`    | Build SRPM + submit to Fedora COPR             |
| `ppa`     | Tag + build and upload to Ubuntu PPA           |
| `publish` | All of the above                               |

If only PPA packaging files changed, increment the PPA revision:

    $ ppa/build-ppa.sh 2

## License

[GNU General Public License v3](LICENSE)
