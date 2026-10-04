#!/bin/sh
set -eu

if [ "$#" -ne 4 ]; then
	echo "usage: smoke-archives.sh <archive-dir> <version> <os> <arch>" >&2
	exit 2
fi

archive_dir=$1
version=$2
os_name=$3
arch_name=$4

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

extract_archive() {
	archive=$1
	dest=$2
	case "$archive" in
		*.zip)
			if command -v unzip >/dev/null 2>&1; then
				unzip -q "$archive" -d "$dest"
			else
				tar -xf "$archive" -C "$dest"
			fi
			;;
		*.tar.gz)
			tar -xzf "$archive" -C "$dest"
			;;
		*)
			echo "unsupported archive: $archive" >&2
			exit 1
			;;
	esac
}

binary_path() {
	name=$1
	base="${name}_${version}_${os_name}_${arch_name}"
	ext=""
	archive="$archive_dir/${base}.tar.gz"
	if [ "$os_name" = "windows" ]; then
		ext=".exe"
		archive="$archive_dir/${base}.zip"
	fi
	if [ ! -f "$archive" ]; then
		echo "missing archive: $archive" >&2
		exit 1
	fi
	extract_archive "$archive" "$tmpdir"
	printf '%s\n' "$tmpdir/$base/${name}${ext}"
}

qc2=$(binary_path qc2)
cpwd=$(binary_path cpwd)

version_out=$("$qc2" version)
version_line=$(printf '%s\n' "$version_out" | head -n 1)
if [ "$version_line" != "qc2 ${version}" ]; then
	echo "unexpected version output: $version_out" >&2
	exit 1
fi

standalone=$("$cpwd" --print)
bundled=$("$qc2" cpwd --print)
if [ "$standalone" != "$bundled" ]; then
	echo "standalone and bundled cpwd output differ" >&2
	echo "standalone: $standalone" >&2
	echo "bundled: $bundled" >&2
	exit 1
fi
if [ -z "$standalone" ]; then
	echo "cpwd --print returned empty output" >&2
	exit 1
fi

echo "smoke test passed for ${os_name}/${arch_name}"
