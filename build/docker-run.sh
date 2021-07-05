#!/usr/bin/env bash
# first you must build docker image, you can use make docker command
# docker build . -f Dockerfile-run -t turingchain-build:latest

sudo docker run -it -p 9671:9671 -p 9672:9672 -p 6060:6060 -p 50051:50051 -l linux-turingchain-run \
    -v "$GOPATH"/src/gitlab.__officeSite__/turingchain/turingchain:/go/src/gitlab.__officeSite__/turingchain/turingchain \
    -w /go/src/gitlab.__officeSite__/turingchain/turingchain turingchain:latest
