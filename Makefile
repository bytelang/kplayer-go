SUB_DIR = proto

.PHONY: build subdirs $(SUB_DIR)

subdirs: $(SUB_DIR)

$(SUB_DIR):
	@+make build-go -C $@

build:
	make subdirs
	CGO_ENABLE=1 \
	go build \
	-gcflags="all=-trimpath=${PWD}" \
	-asmflags "all=-trimpath=${PWD}" \
	-ldflags "-X github.com/bytelang/kplayer/types.MAJOR_TAG=$(shell git describe --tags) \
			  -X github.com/bytelang/kplayer/types.MAJOR_HASH=$(shell git rev-parse --short HEAD) \
			  -X github.com/bytelang/kplayer/types.WebSite=https://kplayer.bytelang.cn" \
	-o build/kplayer
