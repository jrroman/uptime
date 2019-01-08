#!/bin/sh

build_image() {
  docker build -t jrroman/uptime .
}

start_image() {
  docker run \
    --rm \
    -it \
    --name="uptime" \
    --net="host" \
    -v $(pwd)/sites.csv:/sites.csv \
    jrroman/uptime
}

main() {
  build_image
  start_image
}
main
