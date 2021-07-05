#!/bin/bash

version=$(./turingchain -v)
docker build . -f Dockerfile-node -t turingchaincoin/node:"$version"

docker tag turingchaincoin/node:"$version" turingchaincoin/node:latest

docker login
docker push turingchaincoin/node:latest
docker push turingchaincoin/node:"$version"
