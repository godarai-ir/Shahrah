#!/bin/sh
set -eu

/usr/sbin/nft delete table inet lb_public_filter 2>/dev/null || true
exec /usr/sbin/nft -f /etc/nftables.d/lb-public-filter.nft
