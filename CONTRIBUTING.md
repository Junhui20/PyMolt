# Contributing to PyMolt

Thanks for your interest in improving PyMolt! Contributions of all kinds are
welcome — bug reports, feature ideas, docs, and code.

## Before you start

For anything beyond a small fix, please **open an issue first** to discuss the
change so we don't duplicate effort or build something that won't be merged.

## Development setup

**Prerequisites**

- Go 1.26+
- A C compiler (CGo is required by Wails): GCC/Clang, or MinGW on Windows
- Linux only: `libgtk-3-dev` and `libwebkit2gtk-4.0-dev`
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation) (optional, for `wails dev`)

**Build & run**

```bash
git clone https://github.com/Junhui20/PyMolt.git
cd PyMolt
go build -tags desktop,production -ldflags "-s -w" -o pymolt .
./pymolt          # GUI
./pymolt scan     # CLI
```

For live frontend reload during development: `wails dev`.

## Project layout

```
main*.go                 entry point + platform window options
internal/app.go          Wails-bound API surface (the GUI <-> Go bridge)
internal/cli/            CLI commands
internal/detector/       finds Python from each source (one file per source)
internal/analyzer/       health, PATH, packages, uninstall, marketplace, updater
internal/config/         user preferences (scan paths, full-home scan)
internal/models/         shared data types
frontend/dist/index.html the UI (HTML/CSS/JS, embedded into the binary)
```

Platform-specific code uses Go build tags (`_unix.go`, `_windows.go`,
`_darwin.go`, `_linux.go`). Keep shared logic in the untagged file and only put
genuinely OS-specific behavior in the tagged ones.

## Before opening a pull request

```bash
gofmt -w .
go vet ./...
go build -tags desktop,production ./...
```

- Keep changes focused; one logical change per PR.
- Match the surrounding code style; add a doc comment to any new exported symbol.
- Because PyMolt can modify PATH and delete files, be especially careful with any
  change touching `internal/analyzer/uninstall*.go`, `pathfix*.go`, or
  `terminal_*.go`. Never build a shell command by concatenating a filesystem path
  — see the existing `shellQuote`/`cleanArg` helpers.

## Reporting bugs / requesting features

Use the issue templates. For security issues, see [SECURITY.md](SECURITY.md).
