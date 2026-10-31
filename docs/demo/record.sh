#!/bin/bash
# Record docs/demo/demo.gif from demo.tape with vhs, sending for real to a local
# Mailpit SMTP server over STARTTLS. Nothing leaves the machine: Mailpit accepts
# any login, and gmt-mail trusts its throwaway CA through SSL_CERT_FILE (TLS is
# still verified). Needs docker, openssl, go and git.
# Usage: docs/demo/record.sh
set -euo pipefail

DEMO=$(cd "$(dirname "$0")" && pwd)
SRC=$(cd "$DEMO/../.." && pwd)
WORK=$(mktemp -d)
NET="gmt-demo-$$"
cleanup() {
    docker rm -f "$NET-mailpit" >/dev/null 2>&1 || true
    docker network rm "$NET" >/dev/null 2>&1 || true
    rm -rf "$WORK"
}
trap cleanup EXIT

VERSION=$(git -C "$SRC" describe --tags --abbrev=0 | sed 's/^v//')
(cd "$SRC" && CGO_ENABLED=0 go build -trimpath -ldflags "-X main.appVersion=$VERSION \
    -X main.gitCommit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%F)" -o "$WORK/gmt-mail" .)
cp "$DEMO/config.toml" "$DEMO/template.eml" "$DEMO/demo.tape" "$WORK/"

openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes -days 1 -subj "/CN=demo CA" \
    -keyout "$WORK/ca.key" -out "$WORK/ca.pem" 2>/dev/null
openssl req -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes -subj "/CN=mailpit" \
    -keyout "$WORK/smtp.key" -out "$WORK/smtp.csr" 2>/dev/null
openssl x509 -req -in "$WORK/smtp.csr" -CA "$WORK/ca.pem" -CAkey "$WORK/ca.key" -CAcreateserial -days 1 \
    -extfile <(printf 'subjectAltName=DNS:mailpit') -out "$WORK/smtp.pem" 2>/dev/null
chmod 755 "$WORK" "$WORK/gmt-mail" && chmod 644 "$WORK"/*.pem "$WORK"/*.key "$WORK"/*.toml "$WORK"/*.eml "$WORK"/*.tape

docker network create "$NET" >/dev/null
docker run -d --name "$NET-mailpit" --network "$NET" --network-alias mailpit -v "$WORK:/certs:ro" \
    axllent/mailpit --smtp-tls-cert /certs/smtp.pem --smtp-tls-key /certs/smtp.key \
    --smtp-require-starttls --smtp-auth-accept-any >/dev/null
sleep 2

docker run --rm --network "$NET" -v "$WORK:/vhs" -w /vhs \
    -e SSL_CERT_FILE=/vhs/ca.pem -e SMTP_HOST=mailpit -e SMTP_PORT=1025 \
    -e SENDER_EMAIL=frodo@shire.example -e SENDER_PASSWORD=demo \
    ghcr.io/charmbracelet/vhs demo.tape

sent=$(docker exec "$NET-mailpit" wget -qO- http://localhost:8025/api/v1/messages | grep -o '"total":[0-9]*')
echo "mailpit received: $sent"
cp "$WORK/demo.gif" "$DEMO/demo.gif"
echo "wrote $DEMO/demo.gif ($(du -h "$DEMO/demo.gif" | cut -f1))"
