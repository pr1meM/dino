#!/usr/bin/env bash
# Builds dino and installs it to /usr/local/bin so it's available
# system-wide as `dino`.
set -e

cd "$(dirname "$0")"
go build -o dino .

if [ -w /usr/local/bin ]; then
    install -m 755 dino /usr/local/bin/dino
else
    sudo install -m 755 dino /usr/local/bin/dino
fi

echo "Installed to /usr/local/bin/dino"
