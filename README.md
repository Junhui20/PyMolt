# PyMolt

[![Build](https://github.com/Junhui20/PyMolt/actions/workflows/build.yml/badge.svg)](https://github.com/Junhui20/PyMolt/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/Junhui20/PyMolt)](https://github.com/Junhui20/PyMolt/releases)
[![Downloads](https://img.shields.io/github/downloads/Junhui20/PyMolt/total)](https://github.com/Junhui20/PyMolt/releases)
![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)
[![License: MIT](https://img.shields.io/github/license/Junhui20/PyMolt)](LICENSE)

> Scan, fix, and manage all your Python installations in one place.

PyMolt is a lightweight cross-platform desktop app that detects every Python on your system — official installs, uv, pyenv, Homebrew, conda, system packages, and virtual environments — and gives you tools to clean up the mess.

## Why?

Most developers end up with 5+ Python installations from different sources, conflicting PATH entries, orphaned virtual environments eating disk space, and no clear picture of what's actually installed. PyMolt fixes that.

## Features

**Scan & Discover**
- Detects Python from 10+ sources (Official, uv, pyenv, Conda, Homebrew, Chocolatey, Scoop, Microsoft Store, system packages, venvs)
- Detects Python configured in VS Code and PyCharm
- Shows version, architecture, disk size, and source for each installation
- Identifies duplicate versions and orphaned virtual environments

**Health Check**
- Verifies each Python: executable works, pip available, SSL module, site-packages
- Flags broken installations with actionable diagnostics

**PATH Analysis & Repair**
- Visualizes all Python-related PATH entries with priority order
- Detects orphaned entries pointing to deleted installations
- One-click repair to remove dead PATH entries (Windows)

**Package Marketplace**
- Browse 2,600+ curated packages from [awesome-python](https://github.com/dylanhogg/awesome-python)
- Search any package on PyPI in real-time
- Install packages directly into any Python installation
- View, install, uninstall, and export packages per environment

**Python Version Management**
- Install/remove Python versions via [uv](https://github.com/astral-sh/uv)
- Visual overview of all installed and available versions

**Cleanup Tools**
- pip and uv cache size display with one-click clean
- Safe uninstall for any Python source
- Add/remove Python from PATH

**Per-Installation Actions**
- Open terminal with specific Python activated
- Set as default Python
- Export requirements.txt
- View installed packages

## Tech Stack

- **Backend:** Go 1.26
- **Frontend:** HTML/CSS/JS (Wails v2 WebView)
- **Size:** ~11 MB single executable, no install needed
- **Platform:** Windows 10/11, macOS, Linux

## Quick Start

### Download

Download the latest binary from [Releases](https://github.com/Junhui20/PyMolt/releases).

| Platform | File |
|----------|------|
| Windows  | `pymolt-windows.exe` |
| macOS    | `pymolt-macos` |
| Linux (modern) | `pymolt-linux` |
| Linux (legacy) | `pymolt-linux-webkit4.0` |

> [!IMPORTANT]
> **Linux WebKit version.** The GUI links against WebKit2GTK. Most current
> distros (Ubuntu 24.04+, Debian 12+, Fedora, Arch) ship **4.1** — use
> `pymolt-linux`. Older ones (Ubuntu 20.04/22.04) only have **4.0** — use
> `pymolt-linux-webkit4.0`. If the app exits with
> `libwebkit2gtk-4.0.so.37: cannot open shared object file`, you grabbed the
> wrong one. Check what you have with `ldconfig -p | grep webkit2gtk`.

> [!NOTE]
> The binaries are **not yet code-signed**, so your OS will warn on first launch.
> That's expected for a new open-source project — the source is public and builds
> are produced by GitHub Actions.

### First launch

**macOS** — Gatekeeper blocks unsigned apps. Clear the quarantine flag, then run:

```bash
xattr -dr com.apple.quarantine ./pymolt-macos
chmod +x ./pymolt-macos
./pymolt-macos
```

(Or right-click the binary → **Open** → **Open**, or System Settings → Privacy & Security → **Open Anyway**.)

**Windows** — SmartScreen may warn: click **More info** → **Run anyway**.

**Linux** — pick the binary matching your WebKit2GTK version (see the note above):

```bash
chmod +x ./pymolt-linux            # WebKit 4.1 (modern distros)
./pymolt-linux
# …or, on Ubuntu 20.04/22.04:
chmod +x ./pymolt-linux-webkit4.0  # WebKit 4.0 (legacy)
./pymolt-linux-webkit4.0
```

The runtime needs GTK3 and WebKit2GTK installed (they ship with most desktop
environments). If missing: `sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0`
(or `libwebkit2gtk-4.0-37` for the legacy build).

### Verify your download (optional)

Each release publishes `SHA256SUMS.txt`:

```bash
# Linux / macOS
sha256sum -c SHA256SUMS.txt        # or: shasum -a 256 -c SHA256SUMS.txt
```

```powershell
# Windows (PowerShell)
Get-FileHash .\pymolt-windows.exe -Algorithm SHA256
```

### Build from source

```bash
# Prerequisites: Go 1.26+, GCC (for CGo)
# Linux also needs the WebKit/GTK dev headers:
#   modern (4.1): libgtk-3-dev libwebkit2gtk-4.1-dev   -> add tag: webkit2_41
#   legacy (4.0): libgtk-3-dev libwebkit2gtk-4.0-dev   -> no extra tag
git clone https://github.com/Junhui20/PyMolt.git
cd PyMolt
# Modern distros (WebKit 4.1):
CGO_ENABLED=1 go build -tags desktop,production,webkit2_41 -ldflags "-s -w" -o pymolt .
# Legacy distros (WebKit 4.0): drop the webkit2_41 tag.
```

Or, if you have Go set up and just want it on your PATH:

```bash
go install github.com/Junhui20/PyMolt@latest
```

## How It Works

PyMolt scans your system by checking:

**All Platforms:**
- pyenv versions directory
- Conda/Miniconda/Miniforge environments
- uv managed Python installations
- Virtual environments (recursive scan with configurable depth)
- VS Code and PyCharm IDE configurations

**Windows:**
- Windows Registry (`HKLM/HKCU\SOFTWARE\Python`)
- Known installation directories (`C:\Python*`, `AppData\Local\Programs\Python`)
- Chocolatey and Scoop package directories
- Microsoft Store app aliases

**macOS:**
- Homebrew Cellar (`/opt/homebrew/Cellar/python@*`)
- Python.org Framework (`/Library/Frameworks/Python.framework`)
- `/usr/local/bin`, `/usr/bin`

**Linux:**
- System Python (`/usr/bin`, `/usr/local/bin`)
- Deadsnakes PPA versioned binaries
- Snap binaries

All scanning runs in parallel for speed. No data leaves your machine except when you search PyPI or load the package catalog.

## CLI Mode

```
pymolt              Open GUI
pymolt scan         Scan all Python installations
pymolt fix          Auto-detect and fix issues
pymolt versions     List installed/available Python versions
pymolt health       Run health checks
pymolt path         Analyze PATH entries
pymolt cache        Show cache sizes
pymolt help         Show help
```

## Recommended: Use with uv

PyMolt works best with [uv](https://github.com/astral-sh/uv) as your Python version manager. uv can replace pip, pyenv, virtualenv, and more in a single fast binary.

```bash
# Install uv (macOS/Linux)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install uv (Windows)
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Then use PyMolt to manage everything visually
```

## Safety & privacy

PyMolt can edit your PATH and delete Python installations, so it only acts when you
explicitly click an action. Destructive operations (uninstall/delete) are guarded
against protected system and home directories, and your PATH is backed up before a
repair. Note that deleting a venv or installation is **irreversible**. See
[SECURITY.md](SECURITY.md) for exactly what PyMolt changes on your system.

No data leaves your machine except fetching the package catalog, searching PyPI,
and checking for updates. No telemetry.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) and open
an issue first to discuss substantial changes. By participating you agree to the
[Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE)
