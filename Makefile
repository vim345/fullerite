PROG = fullerite
SOURCES += $(wildcard src/fullerite/handlers/*.go)
SOURCES += $(wildcard src/fullerite/collectors/*.go)
SOURCES = $(wildcard src/fullerite/*.go)

all: $(PROG)

deps:
	GOPATH=$(shell pwd) go get $(PROG)

$(PROG): $(SOURCES) deps
	@echo Building $(SOURCES)
	GOPATH=$(shell pwd) go build $@

fmt: $(SOURCES)
	GOPATH=$(shell pwd) go fmt $^

test: tests
tests: deps
	GOPATH=$(shell pwd) go test $(PROG)
