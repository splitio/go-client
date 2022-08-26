# Setup defaults
GO ?= go

sources := $(shell find . -name *.go)
version	:= $(shell cat splitio/version.go | grep 'const Version' | sed 's/const Version = //' | tr -d '"')


.PHONY: help clean test test_norace test_coverage 

default: help

## Print version
version:
	@echo $(version)

## Delete all build files (both final & temporary)
clean:
	rm -f ./coverage.out

## Run the unit tests
test: $(sources) go.sum
	$(GO) test ./... -count=1 -race

## Run the unit tests without race condition detector
test_norace: $(sources) go.sum
	$(GO) test ./... -count=1

### Run unit tests and generate coverage output
test_coverage: $(sources) go.sum
	$(GO) test -v -cover -coverprofile=coverage.out ./...


# Help target borrowed from: https://docs.cloudposse.com/reference/best-practices/make-best-practices/
## This help screen
help:
	@printf "Available targets:\n\n"
	@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
	    helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
		    helpCommand = $$1; \
		    helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
		    gsub("\\\\", "", helpCommand); \
		    gsub(":+$$", "", helpCommand); \
		    printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	    } \
	    { lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
	@printf "\n"
