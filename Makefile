ODIR=./bin
VERSION=`cat VERSION`
GCLOUD_VERSION=`cat VERSION | sed 's/\./-/g'`

all: clean compile

build:
	go build -o bin/ecom-api -ldflags "-X main.version=$(VERSION)" ./cmd/ecom-api/main.go

run:
	go run -ldflags "-X main.version=$(VERSION)" ./cmd/ecom-api/main.go

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ODIR)/ecom-api-alpine-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd/ecom-api/main.go
	GOOS=darwin GOARCH=amd64 go build -o $(ODIR)/ecom-api-darwin-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd/ecom-api/main.go
	GOOS=linux GOARCH=amd64 go build -o $(ODIR)//ecom-api -ldflags "-X main.version=$(VERSION)" ./cmd/ecom-api/main.go

deploy-gae:
	gcloud app deploy --version=$(GCLOUD_VERSION) ./cmd/ecom-api/app.yaml

clean:
	-@rm -r $(ODIR)/* 2> /dev/null || true
