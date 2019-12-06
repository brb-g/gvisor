#!/bin/bash
set -ex

if which use_bazel.sh >/dev/null 2>/dev/null; then
  use_bazel.sh latest
fi
bazel version

readonly KYTHE_VERSION='v0.0.37'
readonly WORKDIR="$(mktemp -d)"
readonly KYTHE_DIR="${WORKDIR}/kythe-${KYTHE_VERSION}"
readonly KZIP_FILENAME="$(git rev-parse HEAD).kzip"

wget -q -O "${WORKDIR}/kythe.tar.gz" \
  "https://github.com/kythe/kythe/releases/download/${KYTHE_VERSION}/kythe-${KYTHE_VERSION}.tar.gz"
tar --no-same-owner -xzf "${WORKDIR}/kythe.tar.gz" --directory "$WORKDIR"

if [[ -n "$KOKORO_ARTIFACTS_DIR" ]]; then
  cd "${KOKORO_ARTIFACTS_DIR}/github/gvisor"
fi
bazel \
  --bazelrc="${KYTHE_DIR}/extractors.bazelrc" \
  build \
  --override_repository kythe_release="${KYTHE_DIR}" \
  --define=kythe_corpus=gvisor.dev \
  //...

"${KYTHE_DIR}/tools/kzip" merge \
  --output "$KZIP_FILENAME" \
  $(find -L bazel-out/*/extra_actions/ -name '*.kzip')
