.PHONY: build build-linux
build:
	CGO_ENABLE=1 go build -o build/kplayer