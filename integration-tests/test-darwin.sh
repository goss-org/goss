#!/usr/bin/env bash

# Install items necessary for integration tests.

# -ET: propagate DEBUG/RETURN/ERR traps to functions and subshells
# -e: exit on unhandled error
# pipefail: any failure in a pipe causes the pipe to fail
set -eET -o pipefail
[[ -n "${DEBUG:-}" ]] && set -x
if ! cd "$(dirname "${BASH_SOURCE[0]}")/.."; then
  echo -e "Failed to cd to repository root"
  return 1
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "Not running $0 on non-Darwin host."
  exit 0
fi

os="${1-darwin}"
arch="${2-amd64}"

pushd "integration-tests/goss"
set +e
out="$(OS="darwin" "../../release/goss-${os}-${arch}" --vars "vars.yaml" --gossfile "darwin/goss.yaml" validate --output tap)"
set -e
echo "output:"
echo "${out}"
egrep -q 'Count: 88, Failed: 0' <<<"$out"
popd
