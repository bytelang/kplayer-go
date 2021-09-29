.PHONY: build
build:
	CGO_ENABLE=1 go build \
	-gcflags="all=-trimpath=${PWD}" \
	-asmflags "all=-trimpath=${PWD}" \
	-o build/kplayer