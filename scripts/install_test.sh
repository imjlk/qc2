#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
QC2_INSTALL_SOURCE_ONLY=1
# shellcheck disable=SC1091
. "$root/scripts/install.sh"

fail() {
	echo "FAIL: $1" >&2
	exit 1
}

assert_eq() {
	if [ "$1" != "$2" ]; then
		echo "FAIL: $3: got '$1', want '$2'" >&2
		exit 1
	fi
}

assert_fail() {
	status=0
	output=$(eval "$1" 2>&1) || status=$?
	if [ "$status" -eq 0 ]; then
		fail "$2"
	fi
	case "$output" in
		*"$3"*) ;;
		*) fail "$2: unexpected output '$output'" ;;
	esac
}

VERSION=0.1.0
assert_eq "$(resolve_tag)" "v0.1.0" "numeric version"
VERSION=v0.1.1
assert_eq "$(resolve_tag)" "v0.1.1" "prefixed version"
assert_eq "$(asset_version "v0.1.1")" "0.1.1" "asset version"
assert_eq "$(asset_version "0.1.1")" "0.1.1" "asset version without prefix"

assert_eq "$(os_from_uname Darwin)" "darwin" "Darwin"
assert_eq "$(os_from_uname Linux)" "linux" "Linux"
assert_fail "os_from_uname WindowsNT" "unsupported OS should fail" "unsupported operating system"

assert_eq "$(arch_from_uname x86_64)" "amd64" "x86_64"
assert_eq "$(arch_from_uname amd64)" "amd64" "amd64"
assert_eq "$(arch_from_uname arm64)" "arm64" "arm64"
assert_eq "$(arch_from_uname aarch64)" "arm64" "aarch64"
assert_fail "arch_from_uname i386" "unsupported arch should fail" "unsupported architecture"

assert_eq "$(detect_os)" "$(os_from_uname "$(uname -s)")" "detect_os"
assert_eq "$(detect_arch)" "$(arch_from_uname "$(uname -m)")" "detect_arch"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
archive_name="qc2_0.1.0_linux_amd64.tar.gz"
printf 'hello\n' > "$tmpdir/$archive_name"
hash=$(archive_hash "$tmpdir/$archive_name")
case "$hash" in
	*[!0-9a-f]*|"") fail "archive hash should be lowercase hex" ;;
esac
if [ "${#hash}" -ne 64 ]; then
	fail "archive hash length ${#hash}, want 64"
fi
printf '%s  %s\n' "$hash" "$archive_name" > "$tmpdir/SHA256SUMS"
verify_archive "$tmpdir/$archive_name" "$archive_name" "$tmpdir/SHA256SUMS"

printf 'other\n' > "$tmpdir/bad.tar.gz"
assert_fail "verify_archive '$tmpdir/bad.tar.gz' '$archive_name' '$tmpdir/SHA256SUMS'" \
	"checksum mismatch should fail" "checksum mismatch"
assert_fail "verify_archive '$tmpdir/bad.tar.gz' 'missing.tar.gz' '$tmpdir/SHA256SUMS'" \
	"missing checksum should fail" "checksum not found"

echo "install.sh tests passed"
