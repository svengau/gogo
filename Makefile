SHELL := /bin/bash

VERSION=0.1.1
GO_FILES=gogo.go encryption.go helpers.go version.go

all: help

.PHONY: all

build:
	rm -rf gogo
	rm -rf go.mod
	
	go build $(GO_FILES)

	go mod init main
	go mod tidy

.PHONY: build

install:
	rm -rf gogo

	go install ./
	

.PHONY: install

help:

	@echo 'Usage: make [TARGET]'
	@echo
	@echo '    make build         build gogo'
	@echo '    make install       install deps'
	@echo

.PHONY: help

