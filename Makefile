SUB_DIR = proto server/proto

.PHONY: build subdirs $(SUB_DIR)

subdirs: $(SUB_DIR)

$(SUB_DIR):
	@+make build-go -C $@

build:
	make subdirs
	CGO_ENABLE=1 go build \
	-gcflags="all=-trimpath=${PWD}" \
	-asmflags "all=-trimpath=${PWD}" \
	-o build/kplayer
