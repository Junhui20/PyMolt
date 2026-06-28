# PyMolt Roadmap

Prioritized feature roadmap, derived from a survey of comparable tools (uv, rye,
pdm, pyenv, conda/mamba/pixi, asdf, mise, Microsoft PyManager) and real user pain
(Hacker News, discuss.python.org, Reddit). Each entry lists the user problem, the
files to touch, a backend/frontend sketch, effort (S/M/L), and risks.

Legend: ✅ implemented · 🔜 planned · ⏳ later · ⛔ out of scope

---

## ✅ 1. End-of-life / security badges  — *implemented*

Every detected interpreter shows a chip — **OK** (green) / **SECURITY** (amber) /
**EOL** (red) — based on the upstream support window, so users can see at a glance
that a Python is a security risk.

- **Backend:** `internal/analyzer/eol.go` — `PythonEOL(majorMinor) models.EOLStatus`
  over a bundled `pythonEOL` table; `internal/models/models.go` — `EOLStatus` + an
  `EOL` field on `PythonInstallation`; `internal/app.go` `Scan()` populates it.
- **Frontend:** `eolBadge()` in `index.html`, rendered next to the version.
- **Tests:** `internal/analyzer/eol_test.go`.
- **Next:** refresh the table at runtime from
  [endoflife.date/python](https://endoflife.date/python) (cache it like the package
  catalog), and surface the badge in the CLI `scan`/`versions` output.

---

## ✅ 2. PEP 668 "externally-managed" badge + guided fix  — *implemented*

**Problem:** Debian/Ubuntu system Python and Homebrew Python ship an
`EXTERNALLY-MANAGED` marker; `pip install` then fails with the infamous error, and
users reach for the foot-gun `--break-system-packages`. This is the single
highest-volume recent Python pain.

**Implemented (this branch):** a no-subprocess filesystem check (`pep668.go` +
`pep668_test.go`) populated in `Scan()`; a purple **PEP 668** chip in the install
table and a "create a venv" hint in the side panel. *The original design (a
`python -c` probe) is kept below; the filesystem check was chosen to keep scans
subprocess-free.*

**Plan**
- New `internal/analyzer/pep668.go`:
  ```go
  // IsExternallyManaged reports whether pip into this interpreter is blocked by PEP 668.
  func IsExternallyManaged(executable string) bool {
      ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
      defer cancel()
      cmd := exec.CommandContext(ctx, executable, "-c",
          "import sysconfig,os;print(int(os.path.exists(os.path.join(sysconfig.get_path('stdlib'),'EXTERNALLY-MANAGED'))))")
      hideWindow(cmd)
      out, err := cmd.Output()
      return err == nil && strings.TrimSpace(string(out)) == "1"
  }
  ```
- Add `ExternallyManaged bool` to `models.PythonInstallation` (populate in `Scan()`
  alongside EOL, or lazily in the health check to avoid an extra spawn per scan).
- **Frontend:** badge "pip blocked (PEP 668)"; replace the install CTA for that row
  with **"Create a venv from this Python"** (wired to the existing `CreateVenv`) or
  "install via pipx/uv tool".
- **Risk:** one extra subprocess per interpreter — fold into the consolidated
  version+arch probe (`detector/common.go`) or compute lazily.
- Source: https://peps.python.org/pep-0668/

---

## ✅ 3. conda `clean` + surface the shared package cache  — *implemented*

**Problem:** `GetCacheInfo` only reports pip/uv today, but conda's package cache
(`pkgs_dirs`) is usually the single largest disk hog (multi-GB). Biggest
disk-reclaim win, and a natural extension of the existing cache cleaner.

**Implemented (this branch):** `condacache.go` (+ wiring) — `GetCondaCacheSize`
sums the conda `pkgs` caches (no subprocess) and shows a "conda packages" row in
the Cleanup tab; `CleanCondaCache` runs `conda clean --all -y` (locating conda on
PATH or under a known root). Rolled into "Clean All".

**Plan**
- New `internal/analyzer/condacache.go`:
  ```go
  func GetCondaCacheSize() (int64, string)        // locate pkgs dir, sum size
  func CleanCondaCache(dryRun bool) *UninstallResult // `conda clean --all -y` (+ --json --dry-run preview)
  ```
  Locate conda from the existing conda detector (`internal/detector/conda.go`) or
  `~/.conda/pkgs` / `<prefix>/pkgs`.
- Extend `CacheInfo` in `app.go` with `CondaSize`/`CondaPath`; add a `CleanCondaCache`
  binding.
- **Frontend:** a third cache row + a dry-run "preview reclaimable" before deleting.
- **Risk:** conda may not be on PATH; gate the row on conda being detected.
- Source: https://docs.conda.io/projects/conda/en/latest/commands/clean.html

---

## 🟡 4. First-class Microsoft PyManager / PEP 514 support (Windows)  — *detection implemented; management later*

**Problem:** The python.org `.exe` installer is deprecated (gone for 3.16+); the new
**PyManager** is the official Windows channel. Riding this wave early is high-leverage.

**Implemented (this branch):** vendor-agnostic PEP 514 read — `OfficialDetector`
now enumerates every `Company\Tag` under `SOFTWARE\Python` (HKLM/HKCU/WOW6432Node),
not just `PythonCore` (`pep514_windows.go`). A new `PyManagerDetector`
(`pymanager_windows.go` + `pymanager_stub.go`) claims `%LocalAppData%\Python`
runtimes as a new `PyManager` source. Cross-compiles for Windows; **the registry /
on-disk layout still needs a real Windows machine to confirm.** Still to do: driving
`%LocalAppData%\Python\bin` shim management and stale-registration repair.
(`py uninstall` routing for PyManager installs is now wired; `py install`/`list`
were intentionally skipped — they'd duplicate the existing uv version management.)

**Plan**
- New `internal/detector/pymanager_windows.go`: read the **full PEP 514 registry**
  vendor-agnostically — iterate `HKLM`/`HKCU\SOFTWARE\Python\<Company>\<Tag>` (today
  only `PythonCore` is read) → `InstallPath\` + `DisplayName`; recognize PyManager
  installs under `%LocalAppData%\Python`.
- New analyzer wrapping `py install` / `py uninstall` / `py list --online`; manage the
  shims in `%LocalAppData%\Python\bin`; offer to repair stale PEP 514 registrations.
- **Frontend:** treat PyManager as a first-class source in the Versions tab.
- **Risk:** Windows-only; **must be verified on a real device**; PEP 514 parsing edge
  cases. Build behind the existing `_windows.go` build tags.
- Sources: https://peps.python.org/pep-0773/ · https://github.com/python/pymanager
  · https://peps.python.org/pep-0514/

---

## 🟡 5. "Which Python wins here?" resolver + project-pin scanner  — *pin scanner implemented; resolver later*

**Problem:** "which python is which" is PyMolt's core identity. Users don't know what
`python` resolves to in a given directory, or that a `.python-version` pin points at
an uninstalled interpreter.

**Implemented (this branch):** `internal/detector/pins.go` (+ `pins_test.go`) —
`ScanProjectPins` walks the scan roots (depth-bounded, `skipDirs`-pruned) for
`.python-version` / `.tool-versions` / `mise.toml`, resolves each against detected
installs, and a "Project Python Pins" section in the Tools tab lists them with an
**Install** button for unsatisfied pins. The full PATH/shim precedence resolver
(the L part) is still later — surfacing the existing `Shadowed` flag is the next step.

**Plan (ship the pin scanner first)**
- New `internal/detector/pins.go`:
  ```go
  type ProjectPin struct {
      Dir, File, Pinned, ResolvedExe string
      Installed bool
  }
  // ScanProjectPins walks the configured scan roots for .python-version,
  // .tool-versions, and mise.toml, resolving each pin against detected installs.
  func ScanProjectPins(roots []string, installs []models.PythonInstallation) []ProjectPin
  ```
  Reuse the venv scan roots (`config.VenvScanPaths`) and `skipDirs` pruning.
- **Frontend:** a "Projects" view listing each pin, its governing directory, the
  resolved interpreter, and a one-click **Install** when the pin is unsatisfied.
- **Then the resolver:** surface the **`Shadowed`** flag already computed in
  `pathanalysis_*.go` ("entry N is shadowed by entry M") prominently — that already
  answers most of "which wins."
- **Risk:** full PATH+pyenv-shim precedence modeling is the L part; the pin scanner +
  surfacing `Shadowed` is the M part and delivers most of the value.
- Sources: https://docs.astral.sh/uv/concepts/python-versions/ · https://github.com/pyenv/pyenv

---

## ⏳ Later
- Windows Store app-execution-alias diagnose + one-click disable (the stub is already
  detected in `store_windows.go` — just add the explainer + fix).
- pip repair from the health check (`ensurepip` / `apt install python3-pip` / `get-pip`).
- ✅ Free-threaded (no-GIL) badge — done (detected via `sys._is_gil_enabled` folded
  into the existing version probe; shows a "no-GIL" chip).
- Unified pipx + `uv tool` global-CLI manager.
- Guided "fully remove Anaconda" (strip `conda init` blocks, `auto_activate_base false`).
- Environment clone; Jupyter kernelspec mapping; lockfile (`pylock.toml`) viewer.

## 🔐 Distribution / signing
Release binaries are unsigned (Gatekeeper/SmartScreen warn on first run). Removing
that requires certificates only the owner can obtain — see **[docs/SIGNING.md](SIGNING.md)**
for the exact secrets and CI steps. Follow-up: package macOS as a `.app`/`.dmg` so
the notarization ticket can be stapled (bare binaries can't be).

## ⛔ Out of scope (stay a *front-end* to uv, don't reimplement it)
- Full project/dependency resolution & locking (poetry/pdm/uv-project territory).
- Per-directory env-var management (direnv/mise `[env]`), task runners, launching
  Jupyter/Spyder, or a standalone IDE interpreter picker.
