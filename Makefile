.PHONY: build install


build:
	go build -o protox cmd/protox/*.go
	buf generate

install:
	go build -o ${GOPATH}/bin/protox cmd/protox/*.go