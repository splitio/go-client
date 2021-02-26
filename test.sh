#!/bin/env bash

if [ ! -f "go.mod" ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
    dep ensure
    export GO111MODULE=off
fi

go test ./... -count=1 -race
