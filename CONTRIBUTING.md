# Contributing to qc2

Thanks for helping improve qc2.

## Development setup

The repository uses Go `1.26.1`.

```bash
git clone https://github.com/imjlk/qc2.git
cd qc2
go test ./...
go vet ./...
go build ./...
```

Run the commands without installing them:

```bash
go run ./cmd/cpwd --print
go run ./cmd/qc2 cpwd --print
```

## Adding a utility

1. Put reusable flags and behavior in `internal/commands/<name>`.
2. Add the standalone entry point in `cmd/<name>/main.go`.
3. Register it in `internal/qc2app/app.go`.
4. Add unit tests and a registration test.
5. Add the binary to the release matrix if it should ship independently.
6. Document the command in `README.md` and `CHANGELOG.md`.

Command packages should not import `internal/cli`; `internal/qc2app` owns that integration.

## Pull requests

- Keep each pull request focused on one change.
- Add tests for behavior changes and bug fixes.
- Run `go test ./...`, `go vet ./...`, and `go build ./...` before opening the pull request.
- Update the changelog when the change affects users.

