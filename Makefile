.PHONY: all build worker install clean version shell

BINARY=boite
VERSION?=dev
LDFLAGS=-X github.com/FacileStudio/boite/cmd.version=$(VERSION)

all: build

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

worker:
	go build -o boite-worker ./cmd/boite-worker

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

version:
	@echo $(VERSION)

clean:
	rm -f $(BINARY) boite-worker

shell:
	@echo "Build successful! Use './boite --help' to get started."
	@echo "To install: make install"