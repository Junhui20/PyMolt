# Security Policy

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Instead, report it privately via GitHub's
[**Report a vulnerability**](https://github.com/Junhui20/PyMolt/security/advisories/new)
button (Security → Advisories). We aim to acknowledge reports within 7 days and
to ship a fix or mitigation as quickly as is practical.

When reporting, please include the OS and PyMolt version (`pymolt help` shows the
version), reproduction steps, and the impact you observed.

## Supply chain & binary trust

- PyMolt's release binaries are built by GitHub Actions from tagged commits.
- Releases publish a `SHA256SUMS.txt`. Verify your download before running:
  - **Linux/macOS:** `sha256sum -c SHA256SUMS.txt` (or `shasum -a 256 -c`)
  - **Windows (PowerShell):** `Get-FileHash .\pymolt-windows.exe -Algorithm SHA256`
- The binaries are currently **not code-signed or notarized**, so macOS Gatekeeper
  and Windows SmartScreen will warn on first run. See the README's
  "First launch" section. If you prefer, build from source.

## What PyMolt changes on your system

PyMolt is a management tool, so some actions modify your machine. It only does so
when you explicitly trigger them:

| Action | What it does |
|--------|--------------|
| Set as default / Add to PATH / Repair PATH | Edits your **user** PATH (on Windows, `HKCU\Environment`). PATH is backed up first. |
| Uninstall (official) | Runs the platform uninstaller (Windows uninstaller, `brew`, `choco`, `scoop`, or `uv`). |
| Uninstall (venv/pyenv/fallback) | Deletes the installation directory after a protected-path safety check. **This is irreversible and is not backed up.** |
| Clean caches | Deletes pip/uv cache directories. |
| Install package / Install Python version | Runs `pip install` / `uv python install`. |

PyMolt makes **no network connection** except: fetching the package catalog
(`raw.githubusercontent.com`), searching PyPI (`pypi.org`), and checking for
updates (`api.github.com`). No telemetry, no analytics, no data leaves your
machine otherwise.
