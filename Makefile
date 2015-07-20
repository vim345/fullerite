PROG    := fullerite
SRCDIR  := src
PKGS    := $(PROG) $(PROG)/metric $(PROG)/handler $(PROG)/collector
SOURCES := $(foreach pkg, $(PKGS), $(wildcard $(SRCDIR)/$(pkg)/*.go))

# symlinks confuse go tools, let's not mess with it and use -L
GOPATH  := $(shell pwd -L)
export GOPATH

all: fmt vet lint $(PROG)

.PHONY: clean
clean:
	rm -f $(PROG) bin/$(PROG)
	rm -rf pkg/*/$(PROG)

deps:
	@echo Getting dependencies...
	@go get $(PROG)

$(PROG): $(SOURCES) deps
	@echo Building $(PROG)...
	@go build $@

test: tests
tests: deps
	@echo Testing $(PROG)
	@go test $(PROG)

fmt: $(SOURCES)
	@$(foreach pkg, $(PKGS), go fmt $(pkg);)

vet: $(SOURCES)
	@echo Vetting $(PROG) sources...
	@go get golang.org/x/tools/cmd/vet
	@$(foreach pkg, $(PKGS), go vet $(pkg);)

lint: $(SOURCES)
	@echo Linting $(PROG) sources...
	@go get github.com/golang/lint/golint
	@$(foreach pkg, $(PKGS), bin/golint $(pkg);)

cyclo: $(SOURCES)
	@echo Checking code complexity...
	@go get github.com/fzipp/gocyclo
	@bin/gocyclo $(SOURCES)
