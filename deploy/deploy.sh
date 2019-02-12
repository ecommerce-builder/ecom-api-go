#!/bin/bash
if [ -z "$VERSION" ]
then
	echo "Please set \$VERSION"
    exit 1
fi

env GOOS=linux GARCH=amd64 CGO_ENABLED=0 \
go build -ldflags "-X main.version=$VERSION" -o ./bin/alpine_amd64/ecom-api-go-alpine-amd64 ./cmd/ecom-api
cp ./bin/alpine_amd64/ecom-api-go-alpine-amd64 ./bin/alpine_amd64/ecom-api-go-alpine-amd64-$VERSION

env GOOS=darwin GARCH=amd64 \
go build -ldflags "-X main.version=$VERSION" -o bin/darwin_amd64/ecom-api-darwin-amd64-$VERSION ./cmd/ecom-api
go build -ldflags "-X main.version=$VERSION" -o bin/linux_amd64/ecom-api-linux-amd64-$VERSION ./cmd/ecom-api

docker build -t gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION .
docker push gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION
docker tag gcr.io/spycameracctv-d48ac/ecom-api-go:$VERSION gcr.io/spycameracctv-d48ac/ecom-api-go:latest
docker push gcr.io/spycameracctv-d48ac/ecom-api-go:latest
