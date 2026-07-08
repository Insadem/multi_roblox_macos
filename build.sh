#!/usr/bin/env bash
#
# build.sh — build multi-roblox-macos for the current architecture and
# place the binary inside multiroblox.app/Contents/MacOS.
#
# Usage:
#   ./build.sh            # build for the current Mac (arm64 or amd64)
#   ./build.sh universal  # build a universal (arm64+amd64) bundle
#
# Requirements: Go 1.23+, Xcode command-line tools (for CGO against
# AppKit / Launch Services).
set -euo pipefail

cd "$(dirname "$0")"

OUT_DIR="multiroblox.app/Contents/MacOS"
BIN_NAME="multi-roblox-macos"
BIN_PATH="${OUT_DIR}/${BIN_NAME}"

mkdir -p "${OUT_DIR}"

if [[ "${1:-}" == "universal" ]]; then
  echo "Building universal (arm64 + amd64) bundle..."
  CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o "${BIN_PATH}.arm64" .
  CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o "${BIN_PATH}.amd64" .
  lipo -create \
    -output "${BIN_PATH}" \
    "${BIN_PATH}.arm64" \
    "${BIN_PATH}.amd64"
  rm -f "${BIN_PATH}.arm64" "${BIN_PATH}.amd64"
else
  echo "Building for $(go env GOARCH)..."
  CGO_ENABLED=1 go build -o "${BIN_PATH}" .
fi

echo "Built ${BIN_PATH}"
echo ""
echo "To run:  open multiroblox.app"
