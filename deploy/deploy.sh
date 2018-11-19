#!/bin/bash

env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o ecom-api-go-alpine cmd/app/main.go

docker build -t gcr.io/spycameracctv-d48ac/ecom-api-go:latest .

docker push gcr.io/spycameracctv-d48ac/ecom-api-go:latest

# kubectl apply -f pod.yaml
