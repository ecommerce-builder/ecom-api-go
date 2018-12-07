#!/bin/bash

export VERSION=0.10.2

env GOOS=linux GARCH=amd64 CGO_ENABLED=0 \
go build -ldflags "-X main.version=$VERSION" -o ecom-api-go-alpine cmd/app/main.go

env GOOS=darwin GARCH=amd64 \
go build -ldflags "-X main.version=$VERSION" -o ecom-api-darwin cmd/app/main.go

go build -ldflags "-X main.version=$VERSION" -o ecom-api-linux cmd/app/main.go

docker build -t gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION .

docker push gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION
docker tag gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION gcr.io/spycameracctv-d48ac/ecom-api-go:latest

docker push gcr.io/spycameracctv-d48ac/ecom-api-go:latest

# kubectl apply -f pod.yaml
