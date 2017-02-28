#!/usr/bin/env bash
# Usage: scripts/archive.sh SOURCE ARCHIVE
# It's assumed that this is ran from the project root.
set -e

SOURCE=$1
ARCHIVE=$2

if [[ $ARCHIVE =~ \.zip$ ]]; then
  zip -j -r - "$SOURCE" > "$ARCHIVE"
else
  tar -cjSC "${SOURCE%/*}" "${SOURCE##*/}" > "$ARCHIVE"
fi

echo "Created archive: $ARCHIVE"
