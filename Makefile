all: lint build

gofmt:
	# run gofmt across all code to format in a standard way
	gofmt -w src

govet:
	# run go vet across all code to pick up on common mistakes
	go tool vet src

lint: gofmt govet

build:
	gb build

test-unit:
	gb test

test-integration:
	gb test -tags integration
