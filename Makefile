# 🌟 Make sure to use literal Tab characters for indentation under targets!
BINARY_NAME=SingboxTrayMonitor.exe

TAG ?= dev
HOMEPAGE_URL ?= https://github.com/tsoiisaiah/SingboxTrayMonitor

.PHONY: all run build clean

all: build

run:
	go run .

build:
	go run github.com/akavel/rsrc@latest -manifest manifest.xml -ico assets/online.ico -o rsrc.syso
	@if [ "$(TAG)" = "dev" ]; then \
		echo "Building in local development mode..."; \
		go build -ldflags="-w -s -H=windowsgui" -o $(BINARY_NAME) . ; \
	else \
		echo "Building in release mode for tag: $(TAG)"; \
		go build -ldflags="-w -s -H=windowsgui -X main.appVersion=$(TAG) -X main.homepageURL=$(HOMEPAGE_URL)/releases/tag/$(TAG)" -o $(BINARY_NAME) . ; \
	fi
	@echo "Build successful! Output binary -> $(BINARY_NAME)"

clean:
	if exist $(BINARY_NAME) del $(BINARY_NAME)
	if exist rsrc.syso del rsrc.syso
	go clean
