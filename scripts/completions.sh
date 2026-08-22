#!/bin/sh
#
# Write the bash, zsh and fish completion scripts into completions/.
#
#   sh scripts/completions.sh
#
# goreleaser runs it from before.hooks, so the release archives and the
# .deb/.rpm/.apk packages carry the completions of the revision being
# released. The directory is generated, not committed.
set -eu

cd "$(dirname "$0")/.."

rm -rf completions
mkdir -p completions

for shell in bash zsh fish; do
	go run ./cmd/ht-cli completion "$shell" >"completions/ht-cli.$shell"
done
