 .PHONY: test setup

help:
	echo "setup, fmt, test"

setup:
	go get -v ./...

fmt:
	gofmt -s -w .

test:
	(cd v1 ; go test .) && (cd v2 ; go test .)
