# PyMolt

> Scan, fix, and manage all your Python installations in one place.

PyMolt is a lightweight Windows desktop app that detects every Python on your system — official installs, uv, pyenv, conda, chocolatey, scoop, Microsoft Store, and virtual environments — and gives you tools to clean up the mess.

## Why?

Most Windows Python developers end up with 5+ Python installations from different sources, conflicting PATH entries, orphaned virtual environments eating disk space, and no clear picture of what's actually installed. PyMolt fixes that.

## Features

**Scan & Discover**
- Detects Python from 8 sources (Official, uv, pyenv, Conda, Chocolatey, Scoop, Microsoft Store, venvs)
- Shows version, architecture, disk size, and source for each installation
- Identifies duplicate versions and orphaned virtual environments

**Health Check**
- Verifies each Python: executable works, pip available, SSL module, site-packages
- Flags broken installations with actionable diagnostics

**PATH Analysis & Repair**
- Visualizes all Python-related PATH entries with priority order
- Detects orphaned entries pointing to deleted installations
- One-click repair to remove dead PATH entries

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
- Add/remove Python from Windows PATH

**Per-Installation Actions**
- Open terminal with specific Python activated
- Set as default Python
- Export requirements.txt
- View installed packages

## Tech Stack

- **Backend:** Go 1.26
- **Frontend:** HTML/CSS/JS (Wails v2 WebView)
- **Size:** ~16 MB single executable, no install needed
- **Platform:** Windows 10/11

## Quick Start

### Download

Download the latest `pymolt.exe` from [Releases](https://github.com/Junhui20/PyMolt/releases).

### Build from source

```bash
# Prerequisites: Go 1.26+, GCC (for CGo)
git clone https://github.com/Junhui20/PyMolt.git
cd PyMolt
go mod tidy
CGO_ENABLED=1 go build -tags desktop,production -ldflags "-H windowsgui" -o pymolt.exe .
```

## How It Works

PyMolt scans your system by checking:
- Windows Registry (`HKLM/HKCU\SOFTWARE\Python`)
- Known installation directories (`C:\Python*`, `AppData\Local\Programs\Python`)
- pyenv-win versions directory
- Conda/Miniconda/Miniforge environments
- uv managed Python installations
- Chocolatey and Scoop package directories
- Microsoft Store app aliases
- Virtual environments (recursive scan with configurable depth)

All scanning runs in parallel for speed. No data leaves your machine except when you search PyPI or load the package catalog.

## Recommended: Use with uv

PyMolt works best with [uv](https://github.com/astral-sh/uv) as your Python version manager. uv can replace pip, pyenv, virtualenv, and more in a single fast binary.

```bash
# Install uv
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Then use PyMolt to manage everything visually
```

## Contributing

Contributions welcome! Please open an issue first to discuss what you'd like to change.

## License

[MIT](LICENSE)
