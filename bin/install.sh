#!/bin/sh

pull_image() {
  docker pull jrroman/uptime
}

start_container() {
  docker run \
    --rm \
    -it \
    --name="uptime" \
    --net="host" \
    -v $(pwd)/sites.csv:/sites.csv \
    jrroman/uptime
}

main() {
  pull_image
  start_container
}
main
