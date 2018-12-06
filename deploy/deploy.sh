#!/bin/bash

VERSION=0.9.3

env \
VERSION=0.9.3 \
GOOS=linux \
GARCH=amd64 \
CGO_ENABLED=0 \
go build -ldflags "-X main.version=$VERSION" -o ecom-api-go-alpine cmd/app/main.go

docker build -t gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION .

docker push gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION
docker push gcr.io/spycameracctv-d48ac/ecom-api-go:latest

# kubectl apply -f pod.yaml
