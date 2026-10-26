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
qc2 version
```

Install both `qc2` and standalone `cpwd`:

```bash
curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | QC2_BINARIES="qc2 cpwd" sh
```

The installer chooses a writable directory already on your `PATH`, preferring `~/.local/bin`, `~/bin`, `/usr/local/bin`, and then Homebrew's bin directory. You can run `qc2` immediately in the same terminal; no `source` command or terminal restart is needed. Shell startup files are not changed.

If none of those directories is writable, the installer uses `/usr/local/bin` when it is on `PATH` and asks for administrator access through `sudo`. Run it in an interactive terminal for the password prompt.

Override the destination with an absolute `QC2_INSTALL_DIR`:

```bash
curl -fsSL https://raw.githubusercontent.com/imjlk/qc2/main/scripts/install.sh | QC2_INSTALL_DIR=/usr/local/bin sh
```

For immediate command access, choose a directory already on `PATH`. A custom destination outside `PATH` still works through its full path, such as `~/.local/bin/qc2 version`. Sourcing `~/.zshrc` or restarting the terminal only helps if that file actually adds the destination to `PATH`.

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

You can also download an archive from [GitHub Releases](https://github.com/imjlk/qc2/releases) and verify it against the published `SHA256SUMS` file.

## Usage

Available bundled commands:

```text
qc2 cpwd       Copy the current working directory.
qc2 list       List available commands.
qc2 help       Show general or command-specific help.
qc2 version    Show the installed version.
```

Release builds also print the commit and build time from `qc2 version`.

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
- Windows: the system clipboard, written as Unicode text
- Linux: the first of `wl-copy`, `xclip`, or `xsel` that accepts the copy

When no Linux backend is available, use `--print` or install one of the listed tools.

## Uninstall

Remove the installed binaries from your selected install directory:

```bash
rm -f "$(command -v qc2)"
# If you also installed the standalone command:
rm -f "$(command -v cpwd)"
```

Use `sudo rm` if the installation required administrator access. For older installations outside `PATH`, remove `~/.local/bin/qc2` and `~/.local/bin/cpwd` directly.

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

Version tags such as `v0.1.0` trigger builds for macOS, Linux, and Windows on both AMD64 and ARM64. GitHub Releases receives the prebuilt archives, generated release notes, and `SHA256SUMS`. Until `v1.0.0`, minor versions may contain breaking changes.

## License

[MIT](LICENSE)
