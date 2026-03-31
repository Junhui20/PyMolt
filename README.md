# PyMolt

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
| Linux    | `pymolt-linux` |

### Build from source

```bash
# Prerequisites: Go 1.26+, GCC (for CGo)
# Linux also needs: libgtk-3-dev libwebkit2gtk-4.0-dev
git clone https://github.com/Junhui20/PyMolt.git
cd PyMolt
go mod tidy
CGO_ENABLED=1 go build -tags desktop,production -ldflags "-s -w" -o pymolt .
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

## Contributing

Contributions welcome! Please open an issue first to discuss what you'd like to change.

## License

[MIT](LICENSE)
