
# Copyright 2020 Adam Chalkley
#
# https://github.com/atc0005/golang-writing-web-applications
#
# Licensed under the BSD 3-Clause "New" or "Revised" License. See LICENSE file
# in the project root for full license information.

# References:
#
# https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies
# https://github.com/mapnik/sphinx-docs/blob/master/Makefile
# https://stackoverflow.com/questions/23843106/how-to-set-child-process-environment-variable-in-makefile
# https://stackoverflow.com/questions/3267145/makefile-execute-another-target
# https://unix.stackexchange.com/questions/124386/using-a-make-rule-to-call-another
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
# https://www.gnu.org/software/make/manual/html_node/Recipe-Syntax.html#Recipe-Syntax
# https://www.gnu.org/software/make/manual/html_node/Special-Variables.html#Special-Variables
# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html
# https://gist.github.com/subfuzion/0bd969d08fe0d8b5cc4b23c795854a13
# https://stackoverflow.com/questions/10858261/abort-makefile-if-variable-not-set
# https://stackoverflow.com/questions/38801796/makefile-set-if-variable-is-empty

SHELL := /bin/bash

OUTPUTBASEFILENAME		= websrv

# https://gist.github.com/TheHippo/7e4d9ec4b7ed4c0d7a39839e6800cc16
VERSION 				= $(shell git describe --always --long --dirty)

# The default `go build` process embeds debugging information. Building
# without that debugging information reduces the binary size by around 28%.
BUILDCMD				=	go build -a -ldflags="-s -w -X main.version=${VERSION}"
GOCLEANCMD				=	go clean
GITCLEANCMD				= 	git clean -xfd
CHECKSUMCMD				=	sha256sum -b

.DEFAULT_GOAL := help

  ##########################################################################
  # Targets will not work properly if a file with the same name is ever
  # created in this directory. We explicitly declare our targets to be phony
  # by making them a prerequisite of the special target .PHONY
  ##########################################################################

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: lintinstall
## lintinstall: install common linting tools
# https://github.com/golang/go/issues/30515#issuecomment-582044819
lintinstall:
	@echo "Installing linting tools"

	@export PATH="${PATH}:$(go env GOPATH)/bin"

	@echo "Installing latest stable staticcheck version via go install command ..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest

	@echo Installing latest stable golangci-lint version per official installation script ...
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	golangci-lint --version

	@echo "Finished updating linting tools"

.PHONY: linting
## linting: runs common linting checks
linting:
	@echo "Running linting tools ..."

	@echo "Running go vet ..."
	@go vet -mod=vendor $(shell go list -mod=vendor ./... | grep -v /vendor/)

	@echo "Running golangci-lint ..."
	@golangci-lint run

	@echo "Running staticcheck ..."
	@staticcheck $(shell go list -mod=vendor ./... | grep -v /vendor/)

	@echo "Finished running linting checks"

.PHONY: gotests
## gotests: runs go test recursively, verbosely
gotests:
	@echo "Running go tests ..."
	@go test -mod=vendor ./...
	@echo "Finished running go tests"

.PHONY: goclean
## goclean: removes local build artifacts, temporary files, etc
goclean:
	@echo "Removing object files and cached files ..."
	@$(GOCLEANCMD)
	@echo "Removing any existing release assets"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-linux-386)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-linux-386.sha256)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-linux-amd64)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-linux-amd64.sha256)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-windows-386.exe)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-windows-386.exe.sha256)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-windows-amd64.exe)"
	@rm -vf "$(wildcard ${OUTPUTBASEFILENAME}-*-windows-amd64.exe.sha256)"

.PHONY: clean
## clean: alias for goclean
clean: goclean

.PHONY: gitclean
## gitclean: WARNING - recursively cleans working tree by removing non-versioned files
gitclean:
	@echo "Removing non-versioned files ..."
	@$(GITCLEANCMD)

.PHONY: pristine
## pristine: run goclean and gitclean to remove local changes
pristine: goclean gitclean

.PHONY: all
# https://stackoverflow.com/questions/3267145/makefile-execute-another-target
## all: generates assets for Linux distros and Windows
all: clean windows linux
	@echo "Completed all cross-platform builds ..."

.PHONY: windows
## windows: generates assets for Windows systems
windows:
	@echo "Building release assets for Windows ..."
	@echo "Building 386 binary"
	@env GOOS=windows GOARCH=386 $(BUILDCMD) -o $(OUTPUTBASEFILENAME)-$(VERSION)-windows-386.exe
	@echo "Building amd64 binary"
	@env GOOS=windows GOARCH=amd64 $(BUILDCMD) -o $(OUTPUTBASEFILENAME)-$(VERSION)-windows-amd64.exe
	@echo "Generating checksum files"
	@$(CHECKSUMCMD) $(OUTPUTBASEFILENAME)-$(VERSION)-windows-386.exe > $(OUTPUTBASEFILENAME)-$(VERSION)-windows-386.exe.sha256
	@$(CHECKSUMCMD) $(OUTPUTBASEFILENAME)-$(VERSION)-windows-amd64.exe > $(OUTPUTBASEFILENAME)-$(VERSION)-windows-amd64.exe.sha256
	@echo "Completed build for Windows"

.PHONY: linux
## linux: generates assets for Linux distros
linux:
	@echo "Building release assets for Linux ..."
	@echo "Building 386 binary"
	@env GOOS=linux GOARCH=386 $(BUILDCMD) -o $(OUTPUTBASEFILENAME)-$(VERSION)-linux-386
	@echo "Building amd64 binary"
	@env GOOS=linux GOARCH=amd64 $(BUILDCMD) -o $(OUTPUTBASEFILENAME)-$(VERSION)-linux-amd64
	@echo "Generating checksum files"
	@$(CHECKSUMCMD) $(OUTPUTBASEFILENAME)-$(VERSION)-linux-386 > $(OUTPUTBASEFILENAME)-$(VERSION)-linux-386.sha256
	@$(CHECKSUMCMD) $(OUTPUTBASEFILENAME)-$(VERSION)-linux-amd64 > $(OUTPUTBASEFILENAME)-$(VERSION)-linux-amd64.sha256
	@echo "Completed build for Linux"
