#!/usr/bin/env bash
# Usage: scripts/build.sh DESTINATION PLATFORM ARCHITECTURE
# It's assumed that this is ran from the project root.
set -e

DEST=$1
PLAT=$2
ARCH=$3
TARGET="${PLAT}/${ARCH}"

declare -A COPY
GO_VERSION="1.8"
BIN_NAME="nmux"
XGO_TMP="_xgotmp"

trap "rm -rf '$XGO_TMP'" EXIT

case $PLAT in
  darwin*)
    app="${DEST}/${PLAT}/${ARCH}/nmux.app/Contents"
    BIN_PATH="${app}/MacOS"
    COPY["assets/macos/Info.plist"]="${app}/Info.plist"
    COPY["assets/macos/AppIcon.icns"]="${app}/Resources/AppIcon.icns"
    ;;
  windows)
    BIN_PATH="${DEST}/windows/${ARCH}"
    BIN_NAME="nmux.exe"
    ;;
  *)
    BIN_PATH="${DEST}/${PLAT}/${ARCH}"
    ;;
esac

[[ ! -d "$BIN_PATH" ]] && mkdir -p "$BIN_PATH"
mkdir -p --mode=2755 "$XGO_TMP"

FINAL_BIN="${BIN_PATH}/${BIN_NAME}"
echo "Building: $FINAL_BIN"
xgo -go "$GO_VERSION" --targets="$TARGET" -dest="$XGO_TMP" ./cmd/nmux >/dev/null || exit 1

cp "${XGO_TMP}/"* "$FINAL_BIN"

for src in "${!COPY[@]}"; do
  dest="${COPY[$src]}"
  mkdir -p "${dest%/*}"
  cp "$src" "$dest"
done
