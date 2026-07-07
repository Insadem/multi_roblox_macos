#!/usr/bin/env bash
#
# build-dmg.sh — package multiroblox.app into a distributable DMG.
#
# Requirements: brew install create-dmg
#
# Usage:
#   ./scripts/build-dmg.sh                      # build an unsigned DMG
#   SIGN_IDENTITY="Developer ID Application: .." \
#     ./scripts/build-dmg.sh                    # build, sign, and notarize
#
# Output:
#   dist/multi-roblox-macos-<version>.dmg
#   dist/multi-roblox-macos-<version>-unsigned.dmg   (if not signing)
#
# Codesign + notarize are skipped automatically if the SIGN_IDENTITY env
# var is empty or unset.
set -euo pipefail

cd "$(dirname "$0")/.."

# Load the version from Info.plist.
VERSION=$(/usr/libexec/PlistBuddy -c "Print :CFBundleShortVersionString" multiroblox.app/Contents/Info.plist)
if [[ -z "$VERSION" ]]; then
  echo "ERROR: could not read CFBundleShortVersionString from multiroblox.app/Contents/Info.plist" >&2
  exit 1
fi

DIST_DIR="dist"
DMG_NAME="multi-roblox-macos-${VERSION}"
APP_NAME="multiroblox.app"
VOL_NAME="multi-roblox-macos ${VERSION}"
FINAL_DMG="${DIST_DIR}/${DMG_NAME}.dmg"
UNSIGNED_DMG="${DIST_DIR}/${DMG_NAME}-unsigned.dmg"

mkdir -p "${DIST_DIR}"
rm -f "${FINAL_DMG}" "${UNSIGNED_DMG}"

# Make sure the app is built and up to date.
./build.sh

# Optional: code-sign the .app before bundling.
if [[ -n "${SIGN_IDENTITY:-}" ]]; then
  echo "Codesigning with identity: ${SIGN_IDENTITY}"
  codesign --force --deep --options runtime --timestamp \
    --sign "${SIGN_IDENTITY}" "${APP_NAME}"
  codesign --verify --verbose=2 "${APP_NAME}"
fi

# Build the DMG with create-dmg. The layout: a window with the app on
# the left, an Applications alias on the right, and a custom background
# (if one is provided in assets/).
echo "Building DMG..."
DMG_BACKGROUND=""
if [[ -f "assets/dmg-background.png" ]]; then
  DMG_BACKGROUND="--background assets/dmg-background.png"
fi

create-dmg \
  --volname "${VOL_NAME}" \
  --window-pos 200 120 \
  --window-size 540 360 \
  --icon-size 96 \
  --icon "${APP_NAME}" 140 170 \
  --hide-extension "${APP_NAME}" \
  --app-drop-link 400 170 \
  --no-internet-enable \
  ${DMG_BACKGROUND} \
  "${UNSIGNED_DMG}" \
  "${APP_NAME}" \
  || true

if [[ ! -f "${UNSIGNED_DMG}" ]]; then
  echo "ERROR: create-dmg did not produce ${UNSIGNED_DMG}" >&2
  exit 1
fi

# Optional: sign and notarize the DMG itself.
if [[ -n "${SIGN_IDENTITY:-}" ]]; then
  echo "Codesigning the DMG..."
  codesign --force --sign "${SIGN_IDENTITY}" "${UNSIGNED_DMG}"

  if [[ -n "${APPLE_ID:-}" && -n "${APPLE_TEAM_ID:-}" && -n "${APPLE_APP_PASSWORD:-}" ]]; then
    echo "Submitting DMG for notarization..."
    xcrun notarytool submit "${UNSIGNED_DMG}" \
      --apple-id "${APPLE_ID}" \
      --team-id "${APPLE_TEAM_ID}" \
      --password "${APPLE_APP_PASSWORD}" \
      --wait
    xcrun stapler staple "${UNSIGNED_DMG}"
  else
    echo "Skipping notarization (APPLE_ID / APPLE_TEAM_ID / APPLE_APP_PASSWORD not set)."
  fi

  mv "${UNSIGNED_DMG}" "${FINAL_DMG}"
  echo "Signed DMG written to ${FINAL_DMG}"
else
  mv "${UNSIGNED_DMG}" "${FINAL_DMG}"
  echo "Unsigned DMG written to ${FINAL_DMG}"
  echo "To sign and notarize, re-run with SIGN_IDENTITY, APPLE_ID, APPLE_TEAM_ID, APPLE_APP_PASSWORD set."
fi

ls -la "${DIST_DIR}"
