#!/usr/bin/env bash
#
# Regenerate the command tree and the reference from the published
# HostTracker API2 v2 specification.
#
#   ./scripts/regen.sh                        # fetch the published spec
#   HT_SPEC=../openapi/openapi-3.0.json ./scripts/regen.sh
#
# It reads the 3.0 twin of the specification, the same document the Go SDK
# is generated from, so the two never disagree about a parameter's shape.
#
# The generated files under cmd/gen are committed: building the CLI must
# not need the specification, and a spec change shows up as a reviewable
# diff of commands rather than as a surprise at release time.
set -euo pipefail

SPEC_URL="${HT_SPEC_URL:-https://raw.githubusercontent.com/HostTracker/openapi/main/openapi-3.0.json}"

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

spec="${HT_SPEC:-}"
cleanup=""
if [ -z "$spec" ]; then
  spec="$(mktemp -t hosttracker-openapi-XXXXXX.json)"
  cleanup="$spec"
  echo "fetching $SPEC_URL"
  curl -fsSL "$SPEC_URL" -o "$spec"
fi
trap '[ -n "$cleanup" ] && rm -f "$cleanup"' EXIT

echo "spec: $spec"
go run ./internal/gen -spec "$spec" -out cmd/gen

go build -o "$(mktemp -d)/ht-cli" ./cmd/ht-cli
go run ./cmd/ht-cli docs --dir docs

gofmt -l cmd internal
go build ./...

echo "regenerated $(ls cmd/gen/*.go | wc -l) files under cmd/gen"
