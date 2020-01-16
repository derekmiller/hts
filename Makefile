.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm hts

build:
	GOOS=linux GOARCH=amd64 go build -o hts