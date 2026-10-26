#!/usr/bin/env sh
set -eu

REPO="${QC2_REPO:-imjlk/qc2}"
INSTALL_DIR="${QC2_INSTALL_DIR:-}"
INSTALL_WITH_SUDO=0
BINARIES="${QC2_BINARIES:-qc2}"
VERSION="${QC2_VERSION:-latest}"

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

require_hash_cmd() {
	if ! command -v sha256sum >/dev/null 2>&1 \
		&& ! command -v shasum >/dev/null 2>&1; then
		echo "missing required command: sha256sum or shasum" >&2
		exit 1
	fi
}

path_contains() (
	# Split PATH without word splitting or glob expansion (paths can contain spaces).
	remaining_path="${PATH:-}"
	while :; do
		path_entry="${remaining_path%%:*}"
		if [ -n "$path_entry" ] && [ "${path_entry%/}" = "${1%/}" ]; then
			return 0
		fi
		case "$remaining_path" in
			*:*) remaining_path="${remaining_path#*:}" ;;
			*) return 1 ;;
		esac
	done
)

can_install_to() (
	# A directory on PATH need not exist yet. Find its nearest existing parent.
	writable_dir="$1"
	while [ ! -e "$writable_dir" ]; do
		writable_dir="${writable_dir%/*}"
		[ -n "$writable_dir" ] || writable_dir=/
	done
	[ -d "$writable_dir" ] && [ -w "$writable_dir" ] && [ -x "$writable_dir" ]
)

select_install_dir() {
	if [ -n "$INSTALL_DIR" ]; then
		case "$INSTALL_DIR" in
			/*) printf '%s\n' "$INSTALL_DIR" ;;
			*) echo "QC2_INSTALL_DIR must be an absolute path" >&2; return 1 ;;
		esac
		return
	fi

	# Use conventional bin directories, never a toolchain's bin or shim directory.
	for candidate in "$HOME/.local/bin" "$HOME/bin" /usr/local/bin /opt/homebrew/bin /home/linuxbrew/.linuxbrew/bin; do
		if path_contains "$candidate" && can_install_to "$candidate"; then
			printf '%s\n' "$candidate"
			return
		fi
	done

	# On a clean macOS/Linux setup, this usually requires administrator access.
	if path_contains /usr/local/bin; then
		printf '%s\n' /usr/local/bin
		return
	fi

	echo "no suitable installation directory is on PATH" >&2
	echo "set QC2_INSTALL_DIR to an absolute bin directory already on PATH" >&2
	return 1
}

install_command() {
	if [ "$INSTALL_WITH_SUDO" = 1 ]; then
		# curl | sh owns stdin; read any sudo password from the terminal instead.
		sudo "$@" </dev/tty
	else
		"$@"
	fi
}

prepare_install_dir() {
	if ! can_install_to "$INSTALL_DIR"; then
		if ! command -v sudo >/dev/null 2>&1 || ! ( : </dev/tty ) 2>/dev/null; then
			echo "cannot write to $INSTALL_DIR; sudo and an interactive terminal are required" >&2
			echo "or set QC2_INSTALL_DIR to a writable bin directory already on PATH" >&2
			exit 1
		fi
		echo "installing to $INSTALL_DIR requires administrator access (sudo)" >&2
		INSTALL_WITH_SUDO=1
	fi
	install_command mkdir -p "$INSTALL_DIR"
}

os_from_uname() {
	case "$1" in
		Darwin) echo "darwin" ;;
		Linux) echo "linux" ;;
		*)
			echo "unsupported operating system" >&2
			return 1
			;;
	esac
}

arch_from_uname() {
	case "$1" in
		x86_64|amd64) echo "amd64" ;;
		arm64|aarch64) echo "arm64" ;;
		*)
			echo "unsupported architecture" >&2
			return 1
			;;
	esac
}

detect_os() {
	os_from_uname "$(uname -s)" || exit 1
}

detect_arch() {
	arch_from_uname "$(uname -m)" || exit 1
}

resolve_tag() {
	if [ "$VERSION" != "latest" ]; then
		case "$VERSION" in
			v*) echo "$VERSION" ;;
			*) echo "v$VERSION" ;;
		esac
		return
	fi

	curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
		| sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' \
		| head -n 1
}

asset_version() {
	echo "$1" | sed 's/^v//'
}

archive_hash() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
		return
	fi
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
		return
	fi

	echo "missing checksum tool: install sha256sum or shasum" >&2
	exit 1
}

verify_archive() {
	archive_path="$1"
	archive_name="$2"
	checksums_path="$3"
	expected="$(awk -v name="$archive_name" '$2 == name {print $1}' "$checksums_path")"

	if [ -z "$expected" ]; then
		echo "checksum not found for $archive_name" >&2
		exit 1
	fi

	actual="$(archive_hash "$archive_path")"
	if [ "$actual" != "$expected" ]; then
		echo "checksum mismatch for $archive_name" >&2
		exit 1
	fi
}

download_binary() {
	name="$1"
	tag="$2"
	os_name="$3"
	arch_name="$4"
	tmpdir="$5"
	checksums_path="$6"
	version_value="$(asset_version "$tag")"
	archive_name="${name}_${version_value}_${os_name}_${arch_name}.tar.gz"
	url="https://github.com/$REPO/releases/download/$tag/$archive_name"

	echo "installing $name from $url"
	if ! curl -fsSL "$url" -o "$tmpdir/$archive_name"; then
		echo "failed to download $archive_name" >&2
		exit 1
	fi
	verify_archive "$tmpdir/$archive_name" "$archive_name" "$checksums_path"
	tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"
	install_command install -m 0755 "$tmpdir/${name}_${version_value}_${os_name}_${arch_name}/$name" "$INSTALL_DIR/$name"
}

main() {
	require_cmd awk
	require_cmd curl
	require_cmd install
	require_cmd sed
	require_cmd tar
	require_hash_cmd

	INSTALL_DIR="$(select_install_dir)"
	OS_NAME="$(detect_os)"
	ARCH_NAME="$(detect_arch)"
	TAG="$(resolve_tag)"

	if [ -z "$TAG" ]; then
		echo "could not resolve a release tag from GitHub" >&2
		exit 1
	fi

	TMPDIR="$(mktemp -d)"
	trap 'rm -rf "$TMPDIR"' EXIT INT TERM

	CHECKSUMS_URL="https://github.com/$REPO/releases/download/$TAG/SHA256SUMS"
	CHECKSUMS_PATH="$TMPDIR/SHA256SUMS"
	if ! curl -fsSL "$CHECKSUMS_URL" -o "$CHECKSUMS_PATH"; then
		echo "failed to download checksums for $TAG: this release may not include SHA256SUMS" >&2
		exit 1
	fi

	prepare_install_dir
	for name in $BINARIES; do
		download_binary "$name" "$TAG" "$OS_NAME" "$ARCH_NAME" "$TMPDIR" "$CHECKSUMS_PATH"
	done

	echo "installed to $INSTALL_DIR"
	if path_contains "$INSTALL_DIR"; then
		for name in $BINARIES; do
			printf 'ready to run: %s --help\n' "$name"
		done
	else
		echo "warning: $INSTALL_DIR is not on PATH; run the binaries by their full path" >&2
		echo "to use bare commands immediately, reinstall with QC2_INSTALL_DIR set to a directory already on PATH" >&2
	fi
}

if [ "${QC2_INSTALL_SOURCE_ONLY:-}" != "1" ]; then
	main
fi
