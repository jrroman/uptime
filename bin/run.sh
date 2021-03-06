#!/bin/sh

set -o errexit

trap "cleanup" EXIT SIGINT SIGTERM SIGKILL

cleanup() {
  make clean
}

build_and_run() {
  make clean
  make install
  make run
}

main() {
  source ./defaults.env
  build_and_run
}
main
