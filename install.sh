#!/bin/sh
#
# Install ht-cli, the official HostTracker command-line client.
#
#   curl -fsSL https://raw.githubusercontent.com/HostTracker/cli/main/install.sh | sh
#
# It reads the latest release of HostTracker/cli, downloads the archive
# built for this platform, checks it against the release's checksums.txt,
# and puts the binary in ~/.local/bin.
#
#   VERSION=v1.2.0   install that release instead of the latest one
#   PREFIX=/usr/local/bin   install somewhere else
#
# Homebrew (brew install HostTracker/tap/ht-cli) and Go (go install
# github.com/HostTracker/cli/cmd/ht-cli@latest) are the other two ways in.
set -eu

REPO="HostTracker/cli"
BINARY="ht-cli"
PREFIX="${PREFIX:-$HOME/.local/bin}"
VERSION="${VERSION:-}"

say() { printf '%s\n' "$*"; }
die() { printf '%s: %s\n' "$BINARY-install" "$*" >&2; exit 1; }

need() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required and was not found"
}

# fetch <url> <destination file>
fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		die "neither curl nor wget is available"
	fi
}

# fetch_stdout <url>
fetch_stdout() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO- "$1"
	else
		die "neither curl nor wget is available"
	fi
}

# platform sets OS and ARCH to the goreleaser names for this machine.
platform() {
	OS="$(uname -s)"
	ARCH="$(uname -m)"
	case "$OS" in
	Linux) OS=linux ;;
	Darwin) OS=darwin ;;
	MINGW* | MSYS* | CYGWIN* | Windows_NT)
		die "this script does not install on Windows; take the zip from https://github.com/$REPO/releases"
		;;
	*) die "unsupported operating system: $OS" ;;
	esac
	case "$ARCH" in
	x86_64 | amd64) ARCH=amd64 ;;
	aarch64 | arm64) ARCH=arm64 ;;
	*) die "unsupported architecture: $ARCH (releases are built for amd64 and arm64)" ;;
	esac
}

# latest_tag reads the tag of the newest published release.
latest_tag() {
	fetch_stdout "https://api.github.com/repos/$REPO/releases/latest" |
		tr ',' '\n' |
		sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
		head -n 1
}

# sha256 prints the checksum of a file, whichever tool this machine has.
sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		die "neither sha256sum nor shasum is available to check the download"
	fi
}

need tar
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 ||
	die "neither sha256sum nor shasum is available to check the download"
platform

if [ -z "$VERSION" ]; then
	VERSION="$(latest_tag)"
	[ -n "$VERSION" ] || die "could not read the latest release of $REPO"
fi
case "$VERSION" in
v*) ;;
*) VERSION="v$VERSION" ;;
esac

# The archive name is goreleaser's name_template: project_version_os_arch,
# with the version stripped of its leading v.
NUMBER="${VERSION#v}"
ARCHIVE="${BINARY}_${NUMBER}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/$REPO/releases/download/$VERSION"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM

say "downloading $ARCHIVE ($VERSION)"
fetch "$BASE/$ARCHIVE" "$TMP/$ARCHIVE" ||
	die "no archive $ARCHIVE in release $VERSION; see https://github.com/$REPO/releases"
fetch "$BASE/checksums.txt" "$TMP/checksums.txt" ||
	die "release $VERSION carries no checksums.txt"

want="$(sed -n "s/^\([0-9a-f]\{64\}\)[[:space:]][[:space:]]*\*\{0,1\}$ARCHIVE\$/\1/p" "$TMP/checksums.txt" | head -n 1)"
[ -n "$want" ] || die "checksums.txt does not list $ARCHIVE"
got="$(sha256 "$TMP/$ARCHIVE")"
[ "$want" = "$got" ] || die "checksum mismatch for $ARCHIVE: expected $want, got $got"
say "checksum ok"

tar -xzf "$TMP/$ARCHIVE" -C "$TMP" "$BINARY" ||
	die "$ARCHIVE does not contain $BINARY"

chmod 0755 "$TMP/$BINARY"
if mkdir -p "$PREFIX" 2>/dev/null && [ -w "$PREFIX" ]; then
	cp "$TMP/$BINARY" "$PREFIX/$BINARY.new" && mv -f "$PREFIX/$BINARY.new" "$PREFIX/$BINARY"
elif command -v sudo >/dev/null 2>&1; then
	say "$PREFIX is not writable; installing with sudo"
	sudo mkdir -p "$PREFIX"
	sudo cp "$TMP/$BINARY" "$PREFIX/$BINARY"
else
	die "$PREFIX is not writable; set PREFIX to a directory you own"
fi

say "installed $BINARY $NUMBER in $PREFIX"

case ":$PATH:" in
*":$PREFIX:"*) ;;
*)
	say ""
	say "$PREFIX is not on your PATH. Add it:"
	say "  export PATH=\"$PREFIX:\$PATH\""
	;;
esac

say "run '$BINARY --help', then '$BINARY auth login'"
