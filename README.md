# qc2

Quick Command Squared

Small CLI utilities for everyday workflows.

`qc2` is a small Go-based CLI toolkit for collecting tiny commands that are useful on their own and still fit together as one bundled command suite.

## Why qc2

The project starts with one Go module and a monorepo-style layout:

- `cmd/<name>` holds standalone binaries such as `cpwd`.
- `cmd/qc2` holds the bundled CLI so users can run `qc2 <command>`.
- `internal/commands/<name>` holds the reusable command logic shared by both entry points.
- `internal/cli` holds the small command registry and built-in help/list/version behavior for the bundled CLI.
- `internal/qc2app` composes the bundled CLI by registering the available commands.

That keeps the first version simple while leaving room to add more utilities like `p2url`, `abspath`, or `mkcd` later.

## Project Structure

```text
cmd/
  cpwd/                 standalone cpwd entry point
  qc2/                  bundled qc2 entry point
internal/
  cli/                  shared command registry and dispatcher
  clip/                 operating-system clipboard backends
  commands/cpwd/        cpwd flags and reusable behavior
  pathutil/             path and file URL helpers
  qc2app/               bundled command composition
  version/              build-time version information
scripts/                release installers
```

The dependency direction is intentional: command packages do not import the bundled CLI. `internal/qc2app` is the composition root that connects commands to `internal/cli`.

## Commands

Current commands:

- `cpwd`: copy the current working directory to the clipboard.

Bundled `qc2` subcommands:

- `qc2 cpwd`
- `qc2 list`
- `qc2 version`
- `qc2 help`

## `cpwd`

Default behavior:

- copy the current working directory to the clipboard
- print the copied value to stdout

Flags:

- `-f`, `--file-url`: copy or print a `file://` URL instead of a plain absolute path
- `-p`, `--print`: print to stdout only and skip the clipboard
- `-q`, `--quiet`: suppress stdout on successful clipboard copy
- `-h`, `--help`: show help

Clipboard backends:

- macOS: `pbcopy`
- Windows: `clip`
- Linux: `wl-copy`, `xclip`, or `xsel`

If no clipboard backend is available, `cpwd` returns a clear error and suggests `--print`.

Examples:

```bash
cpwd
cpwd -f
cpwd -p
cpwd --quiet
cpwd --file-url
cpwd --print
cpwd --file-url --print
```

Bundled usage:

```bash
qc2 cpwd
qc2 cpwd -f
qc2 cpwd --file-url
qc2 cpwd --print
qc2 list
qc2 -v
qc2 version
```

## Local Development

Development expects Go `1.26.1`.

Run tests:

```bash
go test ./...
```

Run static checks:

```bash
go vet ./...
```

Build everything:

```bash
go build ./...
```

Run the standalone command:

```bash
go run ./cmd/cpwd --print
```

Run the bundled CLI:

```bash
go run ./cmd/qc2 cpwd --print
```

## Build Targets

The release workflow is set up to build prebuilt binaries for:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`

Standalone and bundled binaries are archived separately so users can install only `qc2`, only a specific tool like `cpwd`, or both.

## Install Scripts

Draft installers live in:

- `scripts/install.sh` for macOS and Linux
- `scripts/install.ps1` for Windows PowerShell

They are structured around GitHub release assets and can install one or more binaries once releases exist.

## Adding a New Command

1. Add reusable flags and behavior under `internal/commands/<name>`.
2. Add a standalone entry point at `cmd/<name>/main.go`.
3. Register the command in `internal/qc2app/app.go`.
4. Add unit tests for the command and an integration check for registration.
5. Update the README, release workflow, and installer defaults if the new command should ship as a standalone binary.
