build-go:
	protoc -I=.:${PROTO_PATH} --go_out=. --validate_out="lang=go:." --go-grpc_out=. --grpc-gateway_out=logtostderr=true:. *.proto && \
    protoc -I=.:${PROTO_PATH} --grpc-gateway_out=. *.proto
	cp -rf github.com/bytelang/kplayer/types/* ../../types/
	rm -rf github.com

clear:
	rm -rf *.go

