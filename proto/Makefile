SUB_DIR = server config module api

.PHONY: subdirs $(SUB_DIR)

subdirs: $(SUB_DIR)

$(SUB_DIR):
	@+make build-go -C $@

build-go:
	make subdirs
	cd core && \
	protoc -I=. -I=${PROTO_PATH} -I=./proto/prompt -I=./proto/msg --gogo_out=. proto/*.proto && \
	protoc -I=. -I=${PROTO_PATH} -I=./proto/prompt -I=./proto/msg --gogo_out=. proto/prompt/*.proto && \
	protoc -I=. -I=${PROTO_PATH} -I=./proto/prompt -I=./proto/msg --gogo_out=. proto/msg/*.proto && \
	cp -rf github.com/bytelang/kplayer/types/* ../../types/ && \
	rm -rf github.com

clear:
	rm -rf ../types/core ../types/server ../types/config
