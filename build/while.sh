#!/bin/bash

while :; do
    ./turingchain-cli net time
    #nc -vz localhost 9675
    sleep 1
done
