#!/bin/bash
#export CGO_ENABLED=0
#export GOOS=linux
go build -a -installsuffix cgo -o ./build/main cmd/tenkai/*.go
docker build -t softplan/tenkai-api . --no-cache



