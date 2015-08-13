FULLERITE      := fullerite
BEATIT         := beatit
VERSION        := 0.0.7
SRCDIR         := src
HANDLER_DIR    := $(SRCDIR)/fullerite/handler
PROTO_SFX      := $(HANDLER_DIR)/signalfx.proto
GEN_PROTO_SFX  := $(HANDLER_DIR)/signalfx.pb.go
PKGS           := $(FULERITE) $(FULLERITE)/metric $(FULLERITE)/handler $(FULLERITE)/collector $(FULLERITE)/config
SOURCES        := $(foreach pkg, $(PKGS), $(wildcard $(SRCDIR)/$(pkg)/*.go))
SOURCES        := $(filter-out $(GEN_PROTO_SFX), $(SOURCES))



# symlinks confuse go tools, let's not mess with it and use -L
GOPATH  := $(shell pwd -L)
export GOPATH

PATH := bin:$(PATH)
export PATH

all: clean fmt lint $(FULLERITE) $(BEATIT) test

.PHONY: clean
clean:
	@echo Cleaning $(FULLERITE)...
	@rm -f $(FULLERITE) bin/$(FULLERITE)
	@rm -f $(BEATIT) bin/$(BEATIT)
	@rm -rf pkg/*/$(FULLERITE)
	@rm -rf build fullerite*.deb
# Let's keep the generated file in the repo for ease of development.
#	@rm -f $(GEN_PROTO_SFX)

deps:
	@echo Getting dependencies...
	@go get $(FULLERITE)

$(FULLERITE): $(SOURCES) deps
	@echo Building $(FULLERITE)...
	@go build -o bin/$(FULLERITE) $@

$(BEATIT): $(BEATIT_SOURCES)
	@echo Building $(BEATIT)...
	@go build -o bin/$(BEATIT) $@

test: tests
tests: deps
	@echo Testing $(FULLERITE)
	@$(foreach pkg, $(PKGS), go test $(pkg);)

fmt: $(SOURCES)
	@$(foreach pkg, $(PKGS), go fmt $(pkg);)

vet: $(SOURCES)
	@echo Vetting $(FULLERITE) sources...
	@go get -d -u golang.org/x/tools/cmd/vet
	@$(foreach pkg, $(PKGS), go vet $(pkg);)

proto: protobuf
protobuf: $(PROTO_SFX)
	@echo Compiling protobuf
	@go get -u github.com/golang/protobuf/proto
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@protoc --go_out=. $(PROTO_SFX)

lint: $(SOURCES)
	@echo Linting $(FULLERITE) sources...
	@go get -u github.com/golang/lint/golint
	@$(foreach src, $(SOURCES), bin/golint $(src);)

cyclo: $(SOURCES)
	@echo Checking code complexity...
	@go get -u github.com/fzipp/gocyclo
	@bin/gocyclo $(SOURCES)

pkg: package
package: clean $(FULLERITE) $(BEATIT)
	@echo Packaging...
	@mkdir -p build/usr/bin build/usr/share/fullerite build/etc
	@cp bin/fullerite build/usr/bin/
	@cp bin/beatit build/usr/bin/
	@cp bin/run-* build/usr/bin/
	@cp fullerite.conf.example build/etc/
	@cp -r src/diamond build/usr/share/fullerite/diamond
	@fpm -s dir -t deb --name $(FULLERITE) --version $(VERSION) --description "metrics collector" --depends python -C build .
