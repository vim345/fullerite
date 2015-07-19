PROG = fullerite
SOURCES += $(wildcard src/fullerite/handler/*.go)
SOURCES += $(wildcard src/fullerite/collector/*.go)
SOURCES += $(wildcard src/fullerite/*.go)
GO = GOPATH=$(shell pwd) go
all: $(PROG)

deps:
	GOPATH=$(shell pwd) go get $(PROG)

$(PROG): $(SOURCES) deps
	@echo Building $(SOURCES)
	$(GO) build $@

fmt: $(SOURCES)
	$(GO) fmt $^

test: tests
tests: deps
	$(GO) test $(PROG)

cyclo: $(SOURCES)
	$(GO) get github.com/fzipp/gocyclo
	bin/gocyclo $(SOURCES)

lint: $(SOURCES)
	$(GO) get github.com/golang/lint/golint
	bin/golint src/fullerite
	bin/golint src/fullerite/collector
	bin/golint src/fullerite/handler
