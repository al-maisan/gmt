%global goipath github.com/al-maisan/gmt
%global commit  %(git rev-parse --short HEAD 2>/dev/null || echo unknown)

Name:           gmt-mail
Version:        0.2.2
Release:        1%{?dist}
Summary:        Send personalized emails in bulk using templates

License:        GPL-3.0-or-later
URL:            https://%{goipath}
Source0:        https://%{goipath}/archive/v%{version}/gmt-%{version}.tar.gz

BuildRequires:  golang >= 1.25
BuildRequires:  git-core

%description
gmt-mail (Go Mailing Tool) sends personalized emails in bulk. It reads a list
of recipients and mail metadata from an INI configuration file, substitutes
per-recipient variables into an email template, and sends the resulting
messages via SMTP with mandatory TLS encryption.

Features include per-recipient CC and attachment overrides, template variables
for first name, last name, email address, and arbitrary custom fields, a
dry-run mode for previewing emails before sending, and support for file
attachments.

%prep
%autosetup -n gmt-%{version}

%build
go build -mod=vendor -ldflags "-X main.appVersion=%{version} -X main.gitCommit=%{commit} -X main.buildDate=$(date -u +%%Y-%%m-%%d)" -o gmt-mail .

%check
go test -mod=vendor ./...

%install
install -Dpm 0755 gmt-mail %{buildroot}%{_bindir}/gmt-mail
install -Dpm 0644 gmt-mail.1 %{buildroot}%{_mandir}/man1/gmt-mail.1
install -Dpm 0644 .env.example %{buildroot}%{_docdir}/%{name}/.env.example
install -Dpm 0644 README.md %{buildroot}%{_docdir}/%{name}/README.md

%files
%license LICENSE
%{_bindir}/gmt-mail
%{_mandir}/man1/gmt-mail.1*
%{_docdir}/%{name}/

%changelog
* Sat Mar 22 2026 Muharem Hrnjadovic <muharem@linux.com> - 0.2.2-1
- Renamed binary to gmt-mail
- Added man page
- Added PPA and COPR packaging

* Tue Mar 18 2026 Muharem Hrnjadovic <muharem@linux.com> - 0.2.1-1
- Initial package
