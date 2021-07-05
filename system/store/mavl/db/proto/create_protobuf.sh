#!/bin/sh
protoc --go_out=plugins=grpc:../ticket ./*.proto --proto_path=. --proto_path="$GOPATH/src/github.com/turingchain2020/turingchain/types/proto/"
