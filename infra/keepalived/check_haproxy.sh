#!/usr/bin/env sh
set -eu

if pidof haproxy >/dev/null 2>&1; then
  exit 0
fi

exit 1
