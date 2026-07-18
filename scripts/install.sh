#!/usr/bin/env sh
set -eu

REPO="${QC2_REPO:-imjlk/qc2}"
INSTALL_DIR="${QC2_INSTALL_DIR:-$HOME/.local/bin}"
BINARIES="${QC2_BINARIES:-qc2}"
VERSION="${QC2_VERSION:-latest}"

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

detect_os() {
	case "$(uname -s)" in
		Darwin) echo "darwin" ;;
		Linux) echo "linux" ;;
		*)
			echo "unsupported operating system" >&2
			exit 1
			;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64) echo "amd64" ;;
		arm64|aarch64) echo "arm64" ;;
		*)
			echo "unsupported architecture" >&2
			exit 1
			;;
	esac
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
	curl -fsSL "$url" -o "$tmpdir/$archive_name"
	verify_archive "$tmpdir/$archive_name" "$archive_name" "$checksums_path"
	tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"
	install -m 0755 "$tmpdir/${name}_${version_value}_${os_name}_${arch_name}/$name" "$INSTALL_DIR/$name"
}

require_cmd awk
require_cmd curl
require_cmd install
require_cmd sed
require_cmd tar

OS_NAME="$(detect_os)"
ARCH_NAME="$(detect_arch)"
TAG="$(resolve_tag)"

if [ -z "$TAG" ]; then
	echo "could not resolve a release tag from GitHub" >&2
	exit 1
fi

mkdir -p "$INSTALL_DIR"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT INT TERM

CHECKSUMS_URL="https://github.com/$REPO/releases/download/$TAG/SHA256SUMS"
CHECKSUMS_PATH="$TMPDIR/SHA256SUMS"
curl -fsSL "$CHECKSUMS_URL" -o "$CHECKSUMS_PATH"

for name in $BINARIES; do
	download_binary "$name" "$TAG" "$OS_NAME" "$ARCH_NAME" "$TMPDIR" "$CHECKSUMS_PATH"
done

echo "installed to $INSTALL_DIR"
case ":$PATH:" in
	*":$INSTALL_DIR:"*) ;;
	*) echo "add $INSTALL_DIR to PATH to run the installed commands" ;;
esac

