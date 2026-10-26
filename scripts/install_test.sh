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

test_home="$tmpdir/home with spaces"
mkdir -p "$test_home"
assert_eq "$(PATH="$test_home/bin/:/usr/bin:/bin" path_contains "$test_home/bin" && echo yes)" "yes" "PATH with spaces and trailing slash"
assert_fail "PATH='/usr/bin:/bin' path_contains '/usr/bi'" "PATH must match entire entries" ""
assert_fail "PATH='/tmp/[bin]' path_contains '/tmp/b'" "PATH must not expand glob patterns" ""
assert_fail "PATH=':/usr/bin::' path_contains ''" "empty PATH entries are not install destinations" ""

can_install_to "$test_home/new/nested/bin" || fail "missing directories with a writable ancestor should be usable"
assert_fail "can_install_to '$tmpdir/bad.tar.gz'" "a file is not an install directory" ""

INSTALL_DIR=
assert_eq "$(HOME="$test_home" PATH="$test_home/bin:$test_home/.local/bin:/usr/bin:/bin" select_install_dir)" \
	"$test_home/.local/bin" "prefer local bin when it is already on PATH"
assert_eq "$(HOME="$test_home" PATH="$test_home/bin:/usr/bin:/bin" select_install_dir)" \
	"$test_home/bin" "skip local bin when it is absent from PATH"
assert_eq "$(INSTALL_DIR="$test_home/custom bin" PATH=/usr/bin:/bin select_install_dir)" \
	"$test_home/custom bin" "explicit installation destination"
assert_fail "INSTALL_DIR=relative/bin select_install_dir" "relative destinations should fail" "must be an absolute path"
assert_fail "HOME='$test_home' PATH=/usr/bin:/bin select_install_dir" \
	"a default outside PATH should fail" "no suitable installation directory is on PATH"

# Simulate permissions without depending on whether the tests run as root.
selected=$(
	can_install_to() { [ "$1" = /opt/homebrew/bin ]; }
	PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin select_install_dir
)
assert_eq "$selected" /opt/homebrew/bin "prefer writable Homebrew bin over protected local bin"
selected=$(
	can_install_to() { return 1; }
	PATH=/usr/local/bin:/usr/bin:/bin select_install_dir
)
assert_eq "$selected" /usr/local/bin "protected local bin is the sudo fallback"

# Run the actual curl | sh entry point against local release fixtures. The
# calling shell must resolve commands immediately without changing its PATH.
fixture="$tmpdir/release fixtures"
mkdir -p "$fixture/tools" "$fixture/assets" "$fixture/stage"
cat > "$fixture/tools/curl" << 'EOF'
#!/bin/sh
set -eu
destination=
url=
while [ "$#" -gt 0 ]; do
	case "$1" in
		-o) destination=$2; shift 2 ;;
		-*) shift ;;
		*) url=$1; shift ;;
	esac
done
case "$url" in
	*/scripts/install.sh) cat "$QC2_TEST_INSTALLER" ;;
	*/releases/latest) printf '%s\n' '{"tag_name":"v9.9.9"}' ;;
	*/SHA256SUMS|*.tar.gz) cp "$QC2_TEST_ASSETS/${url##*/}" "$destination" ;;
	*) echo "unexpected fixture URL: $url" >&2; exit 1 ;;
esac
EOF
chmod +x "$fixture/tools/curl"

cat > "$fixture/tools/sudo" << 'EOF'
#!/bin/sh
set -eu
[ -t 0 ] || { echo "sudo must read from the terminal" >&2; exit 1; }
if [ "${QC2_TEST_SUDO_DENY:-0}" = 1 ]; then
	echo "sudo fixture: access denied" >&2
	exit 1
fi
exec "$@"
EOF
chmod +x "$fixture/tools/sudo"

status=0
output=$(
	(
		INSTALL_DIR="$tmpdir/admin bin"
		PATH="$fixture/tools:$PATH"
		can_install_to() { return 1; }
		prepare_install_dir || exit $?
		assert_eq "$INSTALL_WITH_SUDO" 1 "protected destination should enable sudo"
		install_command install -m 0755 "$fixture/tools/curl" "$INSTALL_DIR/fixture" || exit $?
	) 2>&1
) || status=$?
if ( : </dev/tty ) 2>/dev/null; then
	assert_eq "$status" 0 "sudo installation from a terminal"
	[ -x "$tmpdir/admin bin/fixture" ] || fail "sudo should install the executable"
	assert_fail "(
		INSTALL_DIR='$tmpdir/denied bin'
		PATH='$fixture/tools':\"\$PATH\"
		QC2_TEST_SUDO_DENY=1
		export QC2_TEST_SUDO_DENY
		can_install_to() { return 1; }
		prepare_install_dir
	)" \
		"sudo denial should fail the installer" "sudo fixture: access denied"
else
	[ "$status" -ne 0 ] || fail "protected installation without a terminal should fail"
	case "$output" in
		*"sudo and an interactive terminal are required"*) ;;
		*) fail "noninteractive permission failure should be actionable" ;;
	esac
	[ ! -e "$tmpdir/admin bin" ] || fail "failed permission check must not create the destination"
fi

: > "$fixture/assets/SHA256SUMS"
for name in qc2 cpwd; do
	base="${name}_9.9.9_$(detect_os)_$(detect_arch)"
	mkdir -p "$fixture/stage/$base"
	cat > "$fixture/stage/$base/$name" << 'EOF'
#!/bin/sh
printf '%s\n' "${0##*/} 9.9.9"
EOF
	chmod +x "$fixture/stage/$base/$name"
	tar -czf "$fixture/assets/$base.tar.gz" -C "$fixture/stage" "$base"
	printf '%s  %s\n' "$(archive_hash "$fixture/assets/$base.tar.gz")" "$base.tar.gz" >> "$fixture/assets/SHA256SUMS"
done

(
	HOME="$tmpdir/default home"
	PATH="$fixture/tools:$HOME/bin:/usr/bin:/bin"
	QC2_TEST_INSTALLER="$root/scripts/install.sh"
	QC2_TEST_ASSETS="$fixture/assets"
	QC2_INSTALL_SOURCE_ONLY=0
	export HOME PATH QC2_TEST_INSTALLER QC2_TEST_ASSETS QC2_INSTALL_SOURCE_ONLY
	unset QC2_INSTALL_DIR QC2_BINARIES QC2_VERSION
	mkdir -p "$HOME/.local/bin"
	printf 'old installation\n' > "$HOME/.local/bin/qc2"
	original_path=$PATH
	curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | sh > "$tmpdir/default.log" 2>&1
	assert_eq "$PATH" "$original_path" "installer must not depend on changing the caller PATH"
	assert_eq "$(command -v qc2)" "$HOME/bin/qc2" "default install resolves immediately"
	assert_eq "$(qc2 version)" "qc2 9.9.9" "default installed binary executes"
	assert_eq "$(sh -c 'qc2 version')" "qc2 9.9.9" "new shell can run the installed binary"
	assert_eq "$(cat "$HOME/.local/bin/qc2")" "old installation" "legacy installation is preserved"
	[ ! -e "$HOME/.zshrc" ] || fail "installer should not create zsh startup files"
	[ ! -e "$HOME/.profile" ] || fail "installer should not create shell profiles"
)

(
	HOME="$tmpdir/local home"
	PATH="$fixture/tools:$HOME/.local/bin/:/usr/bin:/bin"
	QC2_TEST_INSTALLER="$root/scripts/install.sh"
	QC2_TEST_ASSETS="$fixture/assets"
	QC2_INSTALL_SOURCE_ONLY=0
	QC2_BINARIES="qc2 cpwd"
	export HOME PATH QC2_TEST_INSTALLER QC2_TEST_ASSETS QC2_INSTALL_SOURCE_ONLY QC2_BINARIES
	unset QC2_INSTALL_DIR QC2_VERSION
	mkdir -p "$HOME"
	curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | sh > "$tmpdir/local.log" 2>&1
	assert_eq "$(qc2 version)" "qc2 9.9.9" "create missing local bin already on PATH"
	assert_eq "$(cpwd --help)" "cpwd 9.9.9" "standalone binaries resolve immediately"
)

(
	HOME="$tmpdir/custom home"
	PATH="$fixture/tools:/usr/bin:/bin"
	QC2_TEST_INSTALLER="$root/scripts/install.sh"
	QC2_TEST_ASSETS="$fixture/assets"
	QC2_INSTALL_SOURCE_ONLY=0
	QC2_INSTALL_DIR="$HOME/custom bin"
	export HOME PATH QC2_TEST_INSTALLER QC2_TEST_ASSETS QC2_INSTALL_SOURCE_ONLY QC2_INSTALL_DIR
	unset QC2_BINARIES QC2_VERSION
	mkdir -p "$HOME"
	curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | sh > "$tmpdir/custom.log" 2>&1
	assert_eq "$("$QC2_INSTALL_DIR/qc2" version)" "qc2 9.9.9" "custom destination outside PATH is supported"
	case "$(cat "$tmpdir/custom.log")" in
		*"warning: $QC2_INSTALL_DIR is not on PATH"*) ;;
		*) fail "custom destination outside PATH should have a clear warning" ;;
	esac
)

echo "install.sh tests passed"
