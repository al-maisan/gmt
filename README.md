# Go mailing tool (gmt)

`gmt` is a simple utility that sends personalized emails in bulk using a configuration file and a template for the email body. It connects directly via SMTP with mandatory TLS.

## Installation

Requires [Go](https://go.dev/) 1.25 or later.

    $ go build
    $ ./gmt -h

Or install directly:

    $ go install github.com/al-maisan/gmt@latest

## Quick start

Generate sample files, configure SMTP credentials, preview, then send:

    $ ./gmt -sample-config > config.ini
    $ ./gmt -sample-template > template.eml
    $ cp .env.example .env    # edit .env with your SMTP credentials
    $ ./gmt -dry-run -config-path config.ini -template-path template.eml
    $ ./gmt -config-path config.ini -template-path template.eml

## SMTP setup

SMTP credentials are read from environment variables. Copy `.env.example` to `.env` and fill in:

    SMTP_HOST=smtp.gmail.com
    SMTP_PORT=587
    SENDER_EMAIL=your-email@gmail.com
    SENDER_PASSWORD=your-app-password

For Gmail you must use an [App Password](https://myaccount.google.com/apppasswords) (requires 2-Step Verification). Other providers (Outlook, Yahoo, etc.) are documented in `.env.example`.

TLS is enforced -- credentials are never sent in plaintext.

## Configuration file

The config file uses INI format with two sections. Key names are case-insensitive.

### `[general]` section

| Key           | Required | Description                                  |
|---------------|----------|----------------------------------------------|
| `From`        | yes      | Sender name and address for the email header  |
| `subject`     | yes      | Email subject line (supports template vars)   |
| `Cc`          | no       | Comma-separated CC addresses                  |
| `Reply-To`    | no       | Reply-To address                              |
| `attachments` | no       | Comma-separated file paths to attach          |

### `[recipients]` section

Each line defines one recipient:

    email=First Last|KEY:-VALUE|KEY:-VALUE|...

Examples:

    jd@example.com=John Doe Jr.|ORG:-EFF|TITLE:-PhD
    mm@gmail.com=Mickey Mouse|ORG:-Disney

The first word after `=` is the first name, the rest up to the first `|` is the last name. Custom key-value pairs follow, separated by `|`, with `:-` between key and value. Recipients with only a first name (no last name) are also supported.

### Per-recipient overrides

Individual recipients can override the global `Cc` and `attachments`:

    # Replace global Cc entirely
    jd@example.com=John Doe|Cc:-alt@cc.com,other@cc.com

    # Append to global Cc
    daisy@example.com=Daisy Lila|Cc:-+extra@cc.com

    # Replace global attachments
    ab@example.com=Alice Brown|As:-file1.txt,file2.md

    # Append to global attachments
    ef@example.com=Eve Foster|As:-+file3.pdf

A leading `+` means append to the global list; without `+` the global list is replaced.

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

    $ ./gmt -dry-run -config-path config.ini -template-path template.eml
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
| 1    | Missing `-config-path` flag      |
| 2    | Missing `-template-path` flag    |
| 3    | Failed to read config file       |
| 4    | Error parsing config file        |
| 5    | Failed to read template file     |
| 6    | One or more emails failed to send|

## CLI reference

    $ ./gmt -h

    gmt, version 0.2.1
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

## License

[GNU General Public License v3](LICENSE)
