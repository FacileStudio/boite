.PHONY: all build install clean docker-image version shell

BINARY=boite
VERSION?=dev
LDFLAGS=-X github.com/FacileStudio/boite/cmd.Version=$(VERSION)

all: build

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

docker-image:
	docker build -t dev-sandbox:latest .

version:
	@echo $(VERSION)

clean:
	rm -f $(BINARY)

shell:
	@echo "Build successful! Use './boite --help' to get started."
	@echo "To build Docker image: make docker-image"
	@echo "To install: make install"