BINARY_NAME=SingboxTrayMonitor.exe

.PHONY: all run build clean

all: build

run:
	go run .

build:
	go run github.com/akavel/rsrc@latest -manifest manifest.xml -ico assets/online.ico -o rsrc.syso
	go build -ldflags="-w -s -H=windowsgui" -o $(BINARY_NAME) .

clean:
	if exist $(BINARY_NAME) del $(BINARY_NAME)
	if exist rsrc.syso del rsrc.syso
	go clean
