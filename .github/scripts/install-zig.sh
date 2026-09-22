#!/usr/bin/env bash
# Install a pinned, stable Zig release from the official ziglang.org metadata and
# verify it against the published SHA-256 checksum before putting it on PATH.

set -Eeuo pipefail

# Pinned stable release. Source of truth: https://ziglang.org/download/index.json
ZIG_VERSION=0.14.1
# SHA-256 of zig-x86_64-linux-${ZIG_VERSION}.tar.xz from the official index.json.
ZIG_SHA256=24aeeec8af16c381934a6cd7d95c807a8cb2cf7df9fa40d359aa884195c4716c

tarball="zig-x86_64-linux-${ZIG_VERSION}.tar.xz"
url="https://ziglang.org/download/${ZIG_VERSION}/${tarball}"
install_root="${RUNNER_TEMP:?}/zig"

echo "Downloading Zig ${ZIG_VERSION} from ${url}"
curl --fail --location --retry 3 --output "${tarball}" "${url}"

echo "Verifying SHA-256 checksum"
echo "${ZIG_SHA256}  ${tarball}" | sha256sum --check --strict

mkdir -p "${install_root}"
tar -xf "${tarball}" --strip-components 1 -C "${install_root}"
rm -f "${tarball}"

"${install_root}/zig" version

echo "${install_root}" >>"${GITHUB_PATH:?}"
