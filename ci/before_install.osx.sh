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

# stop & disable built-in apache
sudo apachectl stop
sudo launchctl unload -w /System/Library/LaunchDaemons/org.apache.httpd.plist

brew install httpd
# brew's httpd listens on 8080; adjust to make tests pass.
sudo sed -i 's/Listen 8080/Listen 80/' /usr/local/etc/httpd/httpd.conf
sudo brew services start httpd
