#!/usr/bin/env bash
# https://hub.docker.com/r/suyanlong/golang-dev/
# https://github.com/suyanlong/golang-dev
# sudo docker pull suyanlong/golang-dev:latest

sudo docker run -it -p 9671:9671 -p 9672:9672 -p 6060:6060 -p 50051:50051 -l linux-turingchain-build \
    -v "$GOPATH"/src/gitlab.__officeSite__/turingchain/turingchain:/go/src/gitlab.__officeSite__/turingchain/turingchain \
    -w /go/src/gitlab.__officeSite__/turingchain/turingchain suyanlong/golang-dev:latest
