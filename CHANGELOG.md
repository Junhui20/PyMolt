# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project aims
to follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **"Delete venv" and "Remove duplicates" auto-fix buttons did nothing.**
  `ExecuteFix` only implemented `repair_path` and `clean_cache`; every other action
  fell through to a no-op that returned "Use the specific action buttons for this
  fix". The orphaned-venv delete and duplicate-removal actions are now implemented
  (the venv path is carried in the action), they go through the guarded delete
  path (never the OS interpreter), and the summary cards + fix list refresh in
  place afterward. The destructive auto-fixes now ask for confirmation first, like
  the side-panel uninstall.
- **One system Python counted as several "duplicates."** Deduplication keyed on
  the literal executable path, so the same interpreter reached through symlinks —
  `/bin` → `/usr/bin` and `python3` → `python3.X`, common on every modern Linux —
  was reported as 3 separate installs and flagged as a duplicate with a
  **"remove the rest"** button. Following it could have damaged the system
  interpreter. Dedup now resolves symlinks to the real interpreter (venvs stay
  keyed by their own directory), and the cleanup engine never suggests removing an
  OS-managed (`System`) Python.
- A long Version cell (version + EOL + PEP 668 + DEFAULT badges) overflowed into
  the Arch column. The column is wider, the cell now wraps, and `DEFAULT` is a
  proper chip.
- **Linux binary would not launch on modern distros.** Releases were linked
  against WebKit2GTK **4.0**, which Ubuntu 24.04+, Debian 12+, Fedora, and Arch no
  longer ship (they have **4.1**), so the app died at startup with
  `libwebkit2gtk-4.0.so.37: cannot open shared object file`. Releases now publish
  two Linux binaries: `pymolt-linux` (WebKit 4.1, modern) and
  `pymolt-linux-webkit4.0` (legacy). See the README download table for which to use.

### Added
- End-of-life / security badges on every detected interpreter — green **OK**,
  amber **SECURITY**, red **EOL** — based on the upstream support window.
- A **PEP 668** badge marking interpreters whose `pip install` is blocked
  (externally managed), with a side-panel hint to create a virtual environment.
- conda package-cache size in the Cleanup tab, with one-click `conda clean --all`.
- A "Project Python Pins" view (Tools tab): finds `.python-version`,
  `.tool-versions`, and `mise.toml` pins in your project directories and flags any
  that point at a Python you don't have installed.
- Windows: a vendor-agnostic PEP 514 registry read (all `Company\Tag` entries, not
  just `PythonCore`) and detection of Microsoft **PyManager** runtimes under
  `%LocalAppData%\Python` as a new source; uninstall is routed through `py uninstall`.
- A "no-GIL" badge marking free-threaded (PEP 703) interpreters.
- `docs/SIGNING.md`: how to enable macOS/Windows code signing + notarization.
- The project's first unit tests: the delete guard, version comparison, argument
  validation, EOL status, PEP 668 detection, and project-pin scanning.

### Security
- Fixed a shell/AppleScript command-injection path when opening a terminal for an
  installation whose directory name contained shell metacharacters; paths are now
  shell-quoted and AppleScript-escaped.
- Hardened the "uninstall/delete" guard: an allow-list-style check now refuses the
  filesystem root, protected system directories, the home directory, and any
  ancestor of them (previously a small exact-match deny-list missed `/etc`, the
  home directory, `System32`, and others).
- Added a Content-Security-Policy and a quote-safe escaper to the UI; package and
  version inputs to pip/uv are validated to prevent argument injection.
- Terminal launch on Windows now uses a randomly-named temp file instead of a
  predictable shared one.
- The delete guard additionally requires the target directory to actually contain
  a Python interpreter, so a mis-detected or planted folder (e.g. a stray
  `pyvenv.cfg`) can't be turned into a deletable "venv".

### Fixed
- The in-app updater now reports the correct version: the build injects the version
  via `-ldflags -X`, and version comparison is numeric (so `0.10.0` > `0.9.0`).
- The marketplace catalog cache now uses the OS config directory on all platforms
  (previously it wrote a stray relative folder and re-downloaded every launch on
  Linux/macOS). "Latest versions" on PyPI are now sorted correctly.

### Changed
- Branding unified to "PyMolt" across the UI title, header, and HTTP user agents.
- Health checks run concurrently and only when the Health tab is opened.
- Directory size scanning uses `filepath.WalkDir`, and version+architecture are
  detected in a single subprocess per interpreter.
- CI: least-privilege permissions, reproducible builds (`-mod=readonly`), a
  format/vet gate, and `SHA256SUMS.txt` published with each release.

## [0.2.0]

Initial public-preview release: cross-source Python detection, health checks,
PATH analysis & repair, package marketplace, version management via uv, cleanup
tools, and a CLI.
