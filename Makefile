.DEFAULT_GOAL := build

.PHONY:test proto

test:
	gotestsum $(MAKEFLAGS)

test-watch:
	gotestsum --watch $(MAKEFLAGS)

fast-test:
	FAST=true make test

fast-test-watch:
	FAST=true make test-watch

goimports:
	goimports -w .

wsl:
	wsl --fix ./...

fmt:
	gofumpt -w .

fix: goimports fmt wsl

check:
	golangci-lint run

proto:
	sh scripts/compile-protos.sh

vet:
	go vet ./...

get:
	go get ./...
