#!/bin/bash
set -exuo pipefail

go version
#docker build -t go-dns .

echo
sudo ip link add eth10 type dummy
sudo ip address change dev eth10 10.0.0.1
sudo ip link add eth11 type dummy
sudo ip address change dev eth11 10.0.0.2
echo "starting standard http transport test"
if ! go run dns.go; then
#if ! docker run --net=host --rm -it go-dns go run dns.go; then
  echo "failed test"
fi
echo
echo "starting close idle conns test"
if ! go run dns.go --close-idle-conns; then
#if docker run --net=host --rm -it go-dns go run dns.go --close-idle-conns; then
  echo "closing idle connects fixed test"
fi
sudo ip link delete eth10
sudo ip link delete eth11
