#!/usr/bin/env bash
set -e
shopt -s nullglob
GITHUB_TOKEN=$(git config --get github.token)
GITHUB_RELEASE_ARGS="--security-token $GITHUB_TOKEN --user tweekmonster --repo nmux --tag unstable"

git tag -fa -m "Unstable Release" unstable
git push origin -f --tags

github-release delete $GITHUB_RELEASE_ARGS || true
github-release release $GITHUB_RELEASE_ARGS \
  --name "Unstable Build" \
  --description "No guarantee that anything here will work." \
  --pre-release

for archive in dist/*.{zip,tar.bz2}; do
  github-release upload $GITHUB_RELEASE_ARGS \
    --name "${archive##*/}" \
    --file "$archive"
done
