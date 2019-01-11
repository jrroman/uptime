# go commands
GOCMD=go
GOBIN=/go/bin
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# go program binary name
RUN_BIN=uptime

# dep commands
DEPCMD=dep
INIT_DEPS=$(DEPCMD) init
INSTALL_DEPS=$(DEPCMD) ensure


all: clean install_dependencies install
install_dependencies:
	$(INSTALL_DEPS)
	# $(INIT_DEPS)

install: install_dependencies
	$(GOINSTALL) -v ./...

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf \
		$(GOBIN)/$(RUN_BIN) \
		vendor

run:
	$(GOBIN)/$(RUN_BIN) \
		-f "/sites.csv" \
		-d 300s \
		-n 5
