SHELL := /bin/bash

GO_FILES = gogo.go encryption.go helpers.go version.go

VERSION = $(shell ./gogo --version)

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


distribute: build
	@echo 'Build version $(VERSION)'
	rm -rf dist/*
	tar -czf dist/gogo-$(VERSION).tar.gz gogo
	cp gogo dist/
	# shasum is used to publish package with homebrew 
	shasum -a 256 dist/gogo-$(VERSION).tar.gz

.PHONY: distribute

release: distribute
	@echo 'Release version $(VERSION)'
	gh release create v$(VERSION) ./dist/*

.PHONY: release


help:

	@echo 'Usage: make [TARGET]'
	@echo
	@echo '    make build         build gogo'
	@echo '    make install       install deps'
	@echo '    make distribute    build dist'
	@echo

.PHONY: help

