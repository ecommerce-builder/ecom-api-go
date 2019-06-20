ODIR=./bin

all: clean compile 

build:
	go build -o bin/ecom-api ./cmd/ecom-api/main.go

run:
	go run ./cmd/ecom-api/main.go

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ODIR)/ecom-api-alpine-amd64 ./cmd/ecom-api/main.go
	GOOS=darwin GOARCH=amd64 go build -o $(ODIR)/ecom-api-darwin-amd64 ./cmd/ecom-api/main.go
	GOOS=linux GOARCH=amd64 go build -o $(ODIR)//ecom-api ./cmd/ecom-api/main.go

clean:
	-@rm -r $(ODIR)/* 2> /dev/null || true

