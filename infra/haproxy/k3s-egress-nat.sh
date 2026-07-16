#!/bin/sh
set -eu

for src in 10.10.0.21/32 10.10.0.22/32 10.10.0.23/32; do
  iptables -t nat -C POSTROUTING -s "$src" ! -d 10.10.0.0/24 -o eth0 -j MASQUERADE 2>/dev/null \
    || iptables -t nat -A POSTROUTING -s "$src" ! -d 10.10.0.0/24 -o eth0 -j MASQUERADE
done
