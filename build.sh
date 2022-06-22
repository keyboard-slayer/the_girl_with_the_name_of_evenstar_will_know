#!/bin/sh

if [ -d ./build ]; then
    rm -r ./build
fi

cd ./vm/impl/
go build -ldflags="-extldflags=-static" -o ../../build/