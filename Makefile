SUB_DIR = proto

.PHONY: build subdirs $(SUB_DIR)

subdirs: $(SUB_DIR)

$(SUB_DIR):
	@+make build-go -C $@

build:
	make subdirs
	CGO_ENABLE=1 go build \
	-gcflags="all=-trimpath=${PWD}" \
	-asmflags "all=-trimpath=${PWD}" \
	-ldflags "-X github.com/bytelang/kplayer/cmd.MAJOR_TAG=$(shell git describe --tags) \
	          -X github.com/bytelang/kplayer/cmd.WebSite=https://kplayer.bytelang.cn" \
	-o build/kplayer
