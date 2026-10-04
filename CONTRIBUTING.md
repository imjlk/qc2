# Contributing to qc2

Thanks for helping improve qc2.

Participation in this project is governed by the [Code of Conduct](CODE_OF_CONDUCT.md).

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
5. The release workflow ships every directory in `cmd/`. A standalone entry point there is included automatically.
6. Document the command in `README.md` and add a changeset.

Command packages should not import `internal/cli`; `internal/qc2app` owns that integration.

## Pull requests

- Keep each pull request focused on one change.
- Add tests for behavior changes and bug fixes.
- Run `go test ./...`, `go vet ./...`, and `go build ./...` before opening the pull request.
- Add a `.sampo/changesets/*.md` file when the change affects users. The release pull request writes `CHANGELOG.md`.

## Releases

A user-facing changeset looks like this:

```markdown
---
qc2: patch (Fixed)
---

Describe the change for users.
```

The bump is `patch`, `minor`, or `major`. The tag in parentheses is `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, or `Security`. Leave the tag off only when the same sentence is already under `Unreleased`.

After the pull request merges and CI passes, the release prepare workflow opens a release pull request. Merging that pull request tags `vX.Y.Z`. The release workflow then publishes the archives and uses that changelog section as the GitHub release notes.
