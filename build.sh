#!/bin/bash
set -o nounset
set -o errexit

function log() {
    msg=$1
    level=${2:-"INFO"}
    echo "[BUILD] $level: $msg"
}

OUT_DIR=./build
[ ! -d $OUT_DIR ] && mkdir $OUT_DIR

log "building app in docker container"
docker run --rm \
       -v .:/go/src/github.com/schmooser/go-mailer \
       -v "$OUT_DIR":/go/bin \
       --env CGO_ENABLED=0 \
       --env GOOS=linux \
       golang:latest go build -a -installsuffix cgo -v -o /go/bin/mailer
log "app is built successfully"
