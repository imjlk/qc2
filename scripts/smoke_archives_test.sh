#!/bin/sh
set -eu

pack_zip() {
	dest=$1
	parent=$2
	dir=$3
	work=$4
	cat > "$work/packzip.go" << 'EOF'
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: packzip <dest.zip> <parent> <dir>")
		os.Exit(2)
	}
	dest, parent, dir := os.Args[1], os.Args[2], os.Args[3]
	file, err := os.Create(dest)
	if err != nil {
		exitErr(err)
	}

	archive := zip.NewWriter(file)
	root := filepath.Join(parent, dir)
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(parent, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Deflate
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, source)
		source.Close()
		return err
	})
	if err != nil {
		exitErr(err)
	}
	if err := archive.Close(); err != nil {
		exitErr(err)
	}
	if err := file.Close(); err != nil {
		exitErr(err)
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
EOF
	go run "$work/packzip.go" "$dest" "$parent" "$dir"
}

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

os_name=$(uname -s)
case "$os_name" in
	Darwin) os_name=darwin ;;
	Linux) os_name=linux ;;
	MINGW*|MSYS*|CYGWIN*|Windows_NT) os_name=windows ;;
	*)
		echo "unsupported operating system: $os_name" >&2
		exit 1
		;;
esac

arch_name=$(uname -m)
case "$arch_name" in
	x86_64|amd64) arch_name=amd64 ;;
	arm64|aarch64) arch_name=arm64 ;;
	*)
		echo "unsupported architecture: $arch_name" >&2
		exit 1
		;;
esac

version=9.9.9
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
stage=$work/stage
archives=$work/archives
mkdir -p "$stage" "$archives"

ext=""
if [ "$os_name" = "windows" ]; then
	ext=".exe"
fi

ldflags="-X github.com/imjlk/qc2/internal/version.Version=${version}"
for name in qc2 cpwd; do
	base="${name}_${version}_${os_name}_${arch_name}"
	mkdir -p "$stage/$base"
	go build -ldflags "$ldflags" -o "$stage/$base/${name}${ext}" "./cmd/$name"
	cp LICENSE README.md "$stage/$base/"
	if [ "$os_name" = "windows" ]; then
		pack_zip "$archives/$base.zip" "$stage" "$base" "$work"
	else
		tar -C "$stage" -czf "$archives/$base.tar.gz" "$base"
	fi
done

sh "$root/scripts/smoke-archives.sh" "$archives" "$version" "$os_name" "$arch_name"
