PROG       := fullerite
SRCDIR     := src
PROTO_FILE := src/fullerite/handler/signalfx.pb.go
PKGS       := $(PROG) $(PROG)/metric $(PROG)/handler $(PROG)/collector
SOURCES    := $(foreach pkg, $(PKGS), $(wildcard $(SRCDIR)/$(pkg)/*.go))
SOURCES    := $(filter-out $(PROTO_FILE), $(SOURCES))


# symlinks confuse go tools, let's not mess with it and use -L
GOPATH  := $(shell pwd -L)
export GOPATH

PATH := bin:$(PATH)
export PATH

all: clean fmt vet lint protobuf $(PROG)

.PHONY: clean
clean:
	@echo Cleaning $(PROG)...
	@rm -f $(PROG) bin/$(PROG)
	@rm -rf pkg/*/$(PROG)
	@rm -f $(PROTO_FILE)

deps:
	@echo Getting dependencies...
	@go get $(PROG)

$(PROG): $(SOURCES) deps
	@echo Building $(PROG)...
	@go build -o bin/$(PROG) $@

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

protobuf: signalfx.proto
	@go get -u github.com/golang/protobuf/proto
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@protoc --go_out=src/fullerite/handler/ signalfx.proto

lint: $(SOURCES)
	@echo Linting $(PROG) sources...
	@go get github.com/golang/lint/golint
	@$(foreach pkg, $(PKGS), bin/golint $(pkg);)

cyclo: $(SOURCES)
	@echo Checking code complexity...
	@go get github.com/fzipp/gocyclo
	@bin/gocyclo $(SOURCES)
