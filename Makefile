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
	-ldflags "-s -w -X github.com/bytelang/kplayer/types.MAJOR_TAG=$(shell git describe --tags) \
			  -X github.com/bytelang/kplayer/types.MAJOR_HASH=$(shell git rev-parse --short HEAD) \
			  -X github.com/bytelang/kplayer/types.WebSite=${KPLAYER_WEBSITE} \
			  -X github.com/bytelang/kplayer/types.ApiHost=${KPLAYER_API_HOST} \
			  -X github.com/bytelang/kplayer/types.ApiPort=${KPLAYER_API_PORT} \
			  -X github.com/bytelang/kplayer/types.ApiVersion=${KPLAYER_API_VERSION} \
			  -X github.com/bytelang/kplayer/types.CipherKey=${KPLAYER_AES_KEY} \
			  -X github.com/bytelang/kplayer/types.CipherIV=${KPLAYER_AES_IV} \
			  -X github.com/bytelang/kplayer/types.TlsRootCert=${KPLAYER_ROOT_CERT} \
			  -X github.com/bytelang/kplayer/types.TlsClientToken=${KPLAYER_CLIENT_TOKEN}" \
	-o build/kplayer
