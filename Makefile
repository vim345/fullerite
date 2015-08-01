PROG           := fullerite
VERSION        := 0.0.2
SRCDIR         := src
HANDLER_DIR    := $(SRCDIR)/fullerite/handler
PROTO_SFX      := $(HANDLER_DIR)/signalfx.proto
GEN_PROTO_SFX  := $(HANDLER_DIR)/signalfx.pb.go
PKGS           := $(PROG) $(PROG)/metric $(PROG)/handler $(PROG)/collector
SOURCES        := $(foreach pkg, $(PKGS), $(wildcard $(SRCDIR)/$(pkg)/*.go))
SOURCES        := $(filter-out $(GEN_PROTO_SFX), $(SOURCES))


# symlinks confuse go tools, let's not mess with it and use -L
GOPATH  := $(shell pwd -L)
export GOPATH

PATH := bin:$(PATH)
export PATH

all: clean fmt lint $(PROG) test

.PHONY: clean
clean:
	@echo Cleaning $(PROG)...
	@rm -f $(PROG) bin/$(PROG)
	@rm -rf pkg/*/$(PROG)
	@rm -rf build fullerite*.deb
# Let's keep the generated file in the repo for ease of development.
#	@rm -f $(GEN_PROTO_SFX)

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
	@go get -d -u golang.org/x/tools/cmd/vet
	@$(foreach pkg, $(PKGS), go vet $(pkg);)

proto: protobuf
protobuf: $(PROTO_SFX)
	@echo Compiling protobuf
	@go get -u github.com/golang/protobuf/proto
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@protoc --go_out=. $(PROTO_SFX)

lint: $(SOURCES)
	@echo Linting $(PROG) sources...
	@go get -u github.com/golang/lint/golint
	@$(foreach src, $(SOURCES), bin/golint $(src);)

cyclo: $(SOURCES)
	@echo Checking code complexity...
	@go get -u github.com/fzipp/gocyclo
	@bin/gocyclo $(SOURCES)

pkg: package
package: clean $(PROG)
	@echo Packaging...
	@mkdir -p build/usr/bin build/usr/share/fullerite build/etc
	@cp bin/fullerite build/usr/bin/
	@cp bin/run-* build/usr/bin/
	@cp fullerite.conf.example build/etc/
	@cp -r src/diamond build/usr/share/fullerite/diamond
	@fpm -s dir -t deb --name $(PROG) --version $(VERSION) --depends python -C build .
