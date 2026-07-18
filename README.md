# qc2

[![CI](https://github.com/imjlk/qc2/actions/workflows/ci.yml/badge.svg)](https://github.com/imjlk/qc2/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/imjlk/qc2?display_name=tag)](https://github.com/imjlk/qc2/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Quick Command Squared**

Small CLI utilities for everyday workflows.

qc2 ships each utility as a standalone binary and also bundles them under one `qc2 <command>` interface. The project is currently pre-1.0; commands are usable, but interfaces may still evolve between minor releases.

## Install

Prebuilt binaries do not require Go.

### macOS and Linux

Install the bundled `qc2` binary:

```bash
curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | sh
```

Install both `qc2` and standalone `cpwd`:

```bash
curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | QC2_BINARIES="qc2 cpwd" sh
```

The default installation directory is `~/.local/bin`. Override it with `QC2_INSTALL_DIR`.

### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.ps1 | iex
```

The default installation directory is `%USERPROFILE%\AppData\Local\Programs\qc2\bin`.

### Go toolchain

Go users can install commands directly from the module:

```bash
go install github.com/imjlk/qc2/cmd/qc2@latest
go install github.com/imjlk/qc2/cmd/cpwd@latest
```

You can also download an archive directly from [GitHub Releases](https://github.com/imjlk/qc2/releases).

## Usage

Available bundled commands:

```text
qc2 cpwd       Copy the current working directory.
qc2 list       List available commands.
qc2 help       Show general or command-specific help.
qc2 version    Show the installed version.
```

The same `cpwd` implementation is available as a standalone binary:

```bash
cpwd
qc2 cpwd
```

### cpwd

By default, `cpwd` copies the absolute current directory to the clipboard and prints it to stdout.

```text
-f, --file-url  Use a file:// URL instead of a plain path.
-p, --print     Print only; do not access the clipboard.
-q, --quiet     Suppress stdout after a successful copy.
-h, --help      Show command help.
```

Examples:

```bash
cpwd                  # copy and print the absolute path
cpwd -q               # copy without stdout
cpwd -f               # copy and print a file:// URL
qc2 cpwd -f -p        # print a file:// URL without copying
```

Clipboard backends:

- macOS: `pbcopy`
- Windows: `clip`
- Linux: the first available of `wl-copy`, `xclip`, or `xsel`

When no backend is available, use `--print` or install one of the listed Linux tools.

## Uninstall

Remove the installed binaries from your selected install directory:

```bash
rm -f ~/.local/bin/qc2 ~/.local/bin/cpwd
```

## Development

Development uses Go `1.26.1`.

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/cpwd --print
go run ./cmd/qc2 cpwd --print
```

## Project structure

```text
cmd/
  cpwd/                 standalone cpwd entry point
  qc2/                  bundled qc2 entry point
internal/
  cli/                  command registry and dispatcher
  clip/                 operating-system clipboard backends
  commands/cpwd/        cpwd flags and reusable behavior
  pathutil/             path and file URL helpers
  qc2app/               bundled command composition
  version/              build-time version information
scripts/                 release installers
```

Command packages do not depend on the bundled CLI. `internal/qc2app` connects reusable command handlers to `internal/cli`.

See [CONTRIBUTING.md](CONTRIBUTING.md) to add a command or submit a change.

## Release model

Version tags such as `v0.1.0` trigger builds for macOS, Linux, and Windows. GitHub Releases receives the prebuilt archives. Until `v1.0.0`, minor versions may contain breaking changes.

## License

[MIT](LICENSE)
