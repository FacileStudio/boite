.PHONY: all build install clean version shell

BINARY=boite
VERSION?=dev
LDFLAGS=-X github.com/FacileStudio/boite/cmd.Version=$(VERSION)

all: build

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

version:
	@echo $(VERSION)

clean:
	rm -f $(BINARY)

shell:
	@echo "Build successful! Use './boite --help' to get started."
	@echo "To install: make install"