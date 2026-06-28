# Code signing & notarization

PyMolt's release binaries are currently **unsigned**, so on first launch macOS
Gatekeeper and Windows SmartScreen warn the user (the README's "First launch"
section documents the bypass). Removing those warnings requires certificates that
only the project owner can obtain — this guide is everything needed to turn it on.

> **This cannot be done from the codebase alone.** You must obtain the certificates
> below and add them as GitHub repository secrets (Settings → Secrets and variables
> → Actions). Self-signed certs do **not** remove the warnings.

---

## macOS — sign + notarize

**You need (one-time):**
- An **Apple Developer Program** membership ($99/yr).
- A **Developer ID Application** certificate, exported as a `.p12` (base64-encode it:
  `base64 -i cert.p12 | pbcopy`).
- An **App Store Connect API key** (Issuer ID, Key ID, and the `.p8`) for `notarytool`.

**Add these secrets:** `MACOS_CERT_P12_BASE64`, `MACOS_CERT_PASSWORD`,
`APPLE_API_ISSUER_ID`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_P8`.

**Add to the `build-macos` job (after the Build step):**
```yaml
      - name: Sign & notarize
        if: ${{ secrets.MACOS_CERT_P12_BASE64 != '' }}
        env:
          P12: ${{ secrets.MACOS_CERT_P12_BASE64 }}
          P12_PW: ${{ secrets.MACOS_CERT_PASSWORD }}
          API_ISSUER: ${{ secrets.APPLE_API_ISSUER_ID }}
          API_KEY_ID: ${{ secrets.APPLE_API_KEY_ID }}
          API_KEY_P8: ${{ secrets.APPLE_API_KEY_P8 }}
        run: |
          KEYCHAIN=build.keychain
          security create-keychain -p t "$KEYCHAIN"
          security default-keychain -s "$KEYCHAIN"
          security unlock-keychain -p t "$KEYCHAIN"
          echo "$P12" | base64 -d > cert.p12
          security import cert.p12 -k "$KEYCHAIN" -P "$P12_PW" -T /usr/bin/codesign
          security set-key-partition-list -S apple-tool:,apple: -s -k t "$KEYCHAIN"
          IDENTITY=$(security find-identity -v -p codesigning | grep -o '"Developer ID Application.*"' | head -1 | tr -d '"')
          codesign --force --options runtime --timestamp --sign "$IDENTITY" pymolt
          echo "$API_KEY_P8" > key.p8
          ditto -c -k --keepParent pymolt pymolt.zip
          xcrun notarytool submit pymolt.zip --issuer "$API_ISSUER" --key-id "$API_KEY_ID" --key key.p8 --wait
```

> ⚠️ A **bare Mach-O binary cannot be stapled** (`stapler` only works on
> `.app`/`.dmg`/`.pkg`). Signing + notarizing the bare binary still lets Gatekeeper
> verify it *online* on first run. For an offline-clean, double-click `.app` with an
> icon, package the binary into a `.app` bundle (and optionally a `.dmg`) first, then
> sign/notarize/staple **that** — a worthwhile follow-up (roadmap).

---

## Windows — Authenticode

**You need:** a code-signing certificate. A standard OV cert still shows SmartScreen
until it builds reputation; an **EV certificate** or **Azure Trusted Signing** clears
it immediately. (Azure Trusted Signing is the modern, lowest-friction option.)

**Add to the `build-windows` job (after Build), for a `.pfx`:**
```yaml
      - name: Sign
        if: ${{ secrets.WINDOWS_PFX_BASE64 != '' }}
        shell: pwsh
        run: |
          [IO.File]::WriteAllBytes("cert.pfx",[Convert]::FromBase64String($env:WINDOWS_PFX_BASE64))
          & "C:/Program Files (x86)/Windows Kits/10/bin/x64/signtool.exe" sign `
            /f cert.pfx /p $env:WINDOWS_PFX_PASSWORD /fd SHA256 `
            /tr http://timestamp.digicert.com /td SHA256 pymolt.exe
        env:
          WINDOWS_PFX_BASE64: ${{ secrets.WINDOWS_PFX_BASE64 }}
          WINDOWS_PFX_PASSWORD: ${{ secrets.WINDOWS_PFX_PASSWORD }}
```
For **Azure Trusted Signing**, use the `azure/trusted-signing-action` instead (no
cert file — auth via an Azure service principal).

---

## Linux
No signing system to satisfy. Optionally GPG-sign `SHA256SUMS.txt` so users can
verify provenance (`gpg --detach-sign --armor SHA256SUMS.txt`, publish the `.asc`).

---

## Until signing is enabled
The release notes / README already tell users how to get past the first-run warning
(macOS `xattr -dr com.apple.quarantine`, Windows "More info → Run anyway"). That's
the honest interim state.
